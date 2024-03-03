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
	"slices"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func getSampleBaselineStatuses() []FeatureBaselineStatus {
	return []FeatureBaselineStatus{
		{
			FeatureID: "feature1",
			Status:    BaselineStatusUndefined,
			LowDate:   nil,
			HighDate:  nil,
		},
		{
			FeatureID: "feature2",
			Status:    BaselineStatusHigh,
			LowDate:   valuePtr[time.Time](time.Date(2000, time.January, 15, 0, 0, 0, 0, time.UTC)),
			HighDate:  valuePtr[time.Time](time.Date(2000, time.January, 31, 0, 0, 0, 0, time.UTC)),
		},
	}
}

func setupRequiredTablesForBaselineStatus(ctx context.Context,
	client *Client, t *testing.T) {
	sampleFeatures := getSampleFeatures()
	for _, feature := range sampleFeatures {
		err := client.UpsertWebFeature(ctx, feature)
		if err != nil {
			t.Errorf("unexpected error during insert of features. %s", err.Error())
		}
	}
}

// Helper method to get all the statuses in a stable order.
func (c *Client) ReadAllBaselineStatuses(ctx context.Context, _ *testing.T) ([]FeatureBaselineStatus, error) {
	stmt := spanner.NewStatement("SELECT * FROM FeatureBaselineStatus ORDER BY FeatureID ASC")
	iter := c.Single().Query(ctx, stmt)
	defer iter.Stop()

	var ret []FeatureBaselineStatus
	for {
		row, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break // End of results
		}
		if err != nil {
			return nil, errors.Join(ErrInternalQueryFailure, err)
		}
		var status SpannerFeatureBaselineStatus
		if err := row.ToStruct(&status); err != nil {
			return nil, errors.Join(ErrInternalQueryFailure, err)
		}
		ret = append(ret, status.FeatureBaselineStatus)
	}

	return ret, nil
}
func statusEquality(left, right FeatureBaselineStatus) bool {
	return left.FeatureID == right.FeatureID &&
		left.Status == right.Status &&
		((left.LowDate != nil && right.LowDate != nil && left.LowDate.Equal(*right.LowDate)) ||
			left.LowDate == right.LowDate) &&
		((left.HighDate != nil && right.HighDate != nil && left.HighDate.Equal(*right.HighDate)) ||
			left.LowDate == right.LowDate)
}

func TestUpsertFeatureBaselineStatus(t *testing.T) {
	client := getTestDatabase(t)
	ctx := context.Background()
	setupRequiredTablesForBaselineStatus(ctx, client, t)
	sampleStatuses := getSampleBaselineStatuses()

	for _, status := range sampleStatuses {
		err := client.UpsertFeatureBaselineStatus(ctx, status)
		if err != nil {
			t.Errorf("unexpected error during insert. %s", err.Error())
		}
	}

	statuses, err := client.ReadAllBaselineStatuses(ctx, t)
	if err != nil {
		t.Errorf("unexpected error during read all. %s", err.Error())
	}
	if !slices.EqualFunc[[]FeatureBaselineStatus](
		sampleStatuses,
		statuses, statusEquality) {
		t.Errorf("unequal status. expected %+v actual %+v", sampleStatuses, statuses)
	}

	err = client.UpsertFeatureBaselineStatus(ctx, FeatureBaselineStatus{
		FeatureID: "feature1",
		Status:    BaselineStatusHigh,
		LowDate:   valuePtr[time.Time](time.Date(2000, time.February, 15, 0, 0, 0, 0, time.UTC)),
		HighDate:  valuePtr[time.Time](time.Date(2000, time.February, 28, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Errorf("unexpected error during update. %s", err.Error())
	}

	expectedPageAfterUpdate := []FeatureBaselineStatus{
		{
			FeatureID: "feature1",
			Status:    BaselineStatusHigh,
			LowDate:   valuePtr[time.Time](time.Date(2000, time.February, 15, 0, 0, 0, 0, time.UTC)),
			HighDate:  valuePtr[time.Time](time.Date(2000, time.February, 28, 0, 0, 0, 0, time.UTC)),
		},
		{
			FeatureID: "feature2",
			Status:    BaselineStatusHigh,
			LowDate:   valuePtr[time.Time](time.Date(2000, time.January, 15, 0, 0, 0, 0, time.UTC)),
			HighDate:  valuePtr[time.Time](time.Date(2000, time.January, 31, 0, 0, 0, 0, time.UTC)),
		},
	}

	statuses, err = client.ReadAllBaselineStatuses(ctx, t)
	if err != nil {
		t.Errorf("unexpected error during read all after update. %s", err.Error())
	}
	if !slices.EqualFunc[[]FeatureBaselineStatus](
		expectedPageAfterUpdate,
		statuses, statusEquality) {
		t.Errorf("unequal status. expected %+v actual %+v", expectedPageAfterUpdate, statuses)
	}
}