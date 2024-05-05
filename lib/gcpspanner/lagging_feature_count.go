// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcpspanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func init() {
	laggingFeatureCountTemplate = NewQueryTemplate(laggingFeatureCountRawTemplate)
}

// nolint: gochecknoglobals // WONTFIX. Compile the template once at startup. Startup fails if invalid.
var (
	// laggingFeatureCountTemplate is the compiled version of laggingFeatureCountRawTemplate.
	laggingFeatureCountTemplate BaseQueryTemplate
)

// nolint:lll // WONTFIX. Long lines to help with readability.
const laggingFeatureCountRawTemplate = `
SELECT br.ReleaseDate AS ReleaseDate,
       (
           SELECT COUNT(DISTINCT wf.FeatureKey)
           FROM WebFeatures wf
           WHERE
             {{range $param := .OtherBrowserParamNames}}
               EXISTS (
                 SELECT 1
                 FROM BrowserFeatureAvailabilities bfa
                 WHERE bfa.WebFeatureID = wf.ID
                   AND bfa.BrowserName = @{{ $param }}
               )
               AND
             {{end}}
             NOT EXISTS ( -- does not exist in target
               SELECT 1
               FROM BrowserFeatureAvailabilities bfa_target
               WHERE bfa_target.WebFeatureID = wf.ID
                 AND bfa_target.BrowserName = @{{ .TargetBrowserParamName }}
             )
       ) AS Count
FROM (
    SELECT BrowserName, BrowserVersion, ReleaseDate
    FROM BrowserReleases
    WHERE BrowserName IN ({{ range $param := .OtherBrowserParamNames }}@{{ $param }},{{end}} @{{ .TargetBrowserParamName }})
) br
WHERE
  br.BrowserName = @{{ .TargetBrowserParamName }}
  AND ReleaseDate >= @startAt
  AND ReleaseDate <= @endAt
{{if .ReleaseDateParam }}
  AND br.ReleaseDate < @{{ .ReleaseDateParam }}
{{end}}
ORDER BY br.ReleaseDate DESC
LIMIT @limit;
`

type LaggingFeatureCountTemplateData struct {
	TargetBrowserParamName string
	OtherBrowserParamNames []string
	ReleaseDateParam       string
}

// LaggingFeatureCountPage contains the details for the lagging feature support count request.
type LaggingFeatureCountPage struct {
	NextPageToken *string
	Metrics       []LaggingFeatureCount
}

// SpannerLaggingFeatureCount is a wrapper for the lagging feature count that is actually
// stored in spanner. For now, it is the same as LaggingFeatureCount.
type SpannerLaggingFeatureCount struct {
	LaggingFeatureCount
}

// LaggingFeatureCount contains information regarding the count of features implemeneted in all other browsers but not
// in the target browser.
type LaggingFeatureCount struct {
	ReleaseDate time.Time `spanner:"ReleaseDate"`
	Count       int64     `spanner:"Count"`
}

func buildLaggingFeatureCountInBrowserTemplate(
	cursor *LaggingFeatureCountCursor,
	targetBrowser string,
	otherBrowsers []string,
	startAt time.Time,
	endAt time.Time,
	pageSize int,
) spanner.Statement {
	params := map[string]interface{}{}
	targetBrowserParamName := "targetBrowserParam"
	params[targetBrowserParamName] = targetBrowser
	otherBrowsersParamNames := make([]string, 0, len(otherBrowsers))
	for i := range otherBrowsers {
		paramName := fmt.Sprintf("otherBrowser%d", i)
		params[paramName] = otherBrowsers[i]
		otherBrowsersParamNames = append(otherBrowsersParamNames, paramName)
	}
	params["limit"] = pageSize

	releaseDateParamName := ""
	if cursor != nil {
		releaseDateParamName = "releaseDateCursor"
		params[releaseDateParamName] = cursor.ReleaseDate
	}

	params["startAt"] = startAt
	params["endAt"] = endAt

	tmplData := LaggingFeatureCountTemplateData{
		TargetBrowserParamName: targetBrowserParamName,
		OtherBrowserParamNames: otherBrowsersParamNames,
		ReleaseDateParam:       releaseDateParamName,
	}
	sql := laggingFeatureCountTemplate.Execute(tmplData)
	stmt := spanner.NewStatement(sql)
	stmt.Params = params

	return stmt
}

func (c *Client) ListLaggingFeatureCountInBrowser(
	ctx context.Context,
	targetBrowser string,
	otherBrowsers []string,
	startAt time.Time,
	endAt time.Time,
	pageSize int,
	pageToken *string,
) (*LaggingFeatureCountPage, error) {

	var cursor *LaggingFeatureCountCursor
	var err error
	if pageToken != nil {
		cursor, err = decodeLaggingFeatureCountCursor(*pageToken)
		if err != nil {
			return nil, errors.Join(ErrInternalQueryFailure, err)
		}
	}

	txn := c.ReadOnlyTransaction()
	defer txn.Close()

	stmt := buildLaggingFeatureCountInBrowserTemplate(
		cursor,
		targetBrowser,
		otherBrowsers,
		startAt,
		endAt,
		pageSize,
	)

	it := txn.Query(ctx, stmt)
	defer it.Stop()

	var results []LaggingFeatureCount
	for {
		row, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var result SpannerLaggingFeatureCount
		if err := row.ToStruct(&result); err != nil {
			return nil, err
		}
		actualResult := LaggingFeatureCount{
			ReleaseDate: result.ReleaseDate,
			Count:       result.Count,
		}
		results = append(results, actualResult)
	}

	page := LaggingFeatureCountPage{
		Metrics:       results,
		NextPageToken: nil,
	}

	if len(results) == pageSize {
		token := encodeLaggingFeatureCountCursor(results[len(results)-1].ReleaseDate)
		page.NextPageToken = &token
	}

	return &page, nil
}
