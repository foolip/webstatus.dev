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
	"reflect"
	"testing"
	"time"
)

func loadDataForListLaggingFeatureCountInBrowser(ctx context.Context, t *testing.T, client *Client) {
	webFeatures := []WebFeature{
		{FeatureKey: "FeatureX", Name: "Cool API"},
		{FeatureKey: "FeatureY", Name: "Super API"},
		{FeatureKey: "FeatureZ", Name: "Neat API"},
		{FeatureKey: "FeatureW", Name: "Amazing API"},
	}
	for _, feature := range webFeatures {
		err := client.UpsertWebFeature(ctx, feature)
		if err != nil {
			t.Errorf("unexpected error during insert of features. %s", err.Error())
		}
	}

	browserReleases := []BrowserRelease{
		// fooBrowser Releases
		{BrowserName: "fooBrowser", BrowserVersion: "110", ReleaseDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
		{BrowserName: "fooBrowser", BrowserVersion: "111", ReleaseDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
		{BrowserName: "fooBrowser", BrowserVersion: "112", ReleaseDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},

		// barBrowser Releases
		{BrowserName: "barBrowser", BrowserVersion: "113", ReleaseDate: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		{BrowserName: "barBrowser", BrowserVersion: "114", ReleaseDate: time.Date(2024, 3, 28, 0, 0, 0, 0, time.UTC)},

		// bazBrowser Releases
		{BrowserName: "bazBrowser", BrowserVersion: "16.4", ReleaseDate: time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)},
		{BrowserName: "bazBrowser", BrowserVersion: "16.5", ReleaseDate: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)},
	}
	for _, release := range browserReleases {
		err := client.InsertBrowserRelease(ctx, release)
		if err != nil {
			t.Errorf("unexpected error during insert of releases. %s", err.Error())
		}
	}

	browserFeatureAvailabilities := []struct {
		FeatureKey string
		BrowserFeatureAvailability
	}{
		// fooBrowser Availabilities
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "fooBrowser", BrowserVersion: "111"},
			FeatureKey:                 "FeatureX",
		}, // Available from fooBrowser 111
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "fooBrowser", BrowserVersion: "112"},
			FeatureKey:                 "FeatureY",
		}, // Available from fooBrowser 112

		// barBrowser Availabilities
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "barBrowser", BrowserVersion: "113"},
			FeatureKey:                 "FeatureX",
		}, // Available from barBrowser 113
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "barBrowser", BrowserVersion: "113"},
			FeatureKey:                 "FeatureZ",
		}, // Available from barBrowser 113
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "barBrowser", BrowserVersion: "114"},
			FeatureKey:                 "FeatureY",
		}, // Available from barBrowser 114
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "barBrowser", BrowserVersion: "115"},
			FeatureKey:                 "FeatureW",
		}, // Available from barBrowser 115

		// bazBrowser Availabilities
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "bazBrowser", BrowserVersion: "16.4"},
			FeatureKey:                 "FeatureX",
		}, // Available from bazBrowser 111
		{
			BrowserFeatureAvailability: BrowserFeatureAvailability{BrowserName: "bazBrowser", BrowserVersion: "16.5"},
			FeatureKey:                 "FeatureY",
		}, // Available from bazBrowser 112
	}
	for _, availability := range browserFeatureAvailabilities {
		err := client.InsertBrowserFeatureAvailability(ctx,
			availability.FeatureKey, availability.BrowserFeatureAvailability)
		if err != nil {
			t.Errorf("unexpected error during insert. %s", err.Error())
		}
	}
}

func TestListLaggingFeatureCountInBrowser(t *testing.T) {
	client := getTestDatabase(t)
	ctx := context.Background()

	loadDataForListLaggingFeatureCountInBrowser(ctx, t, client)

	t.Run("all data", func(t *testing.T) {
		targetBrowser := "fooBrowser"
		otherBrowsers := []string{
			"barBrowser",
			"bazBrowser",
		}
		startAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		endAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		pageSize := 100

		result, err := client.ListLaggingFeatureCountInBrowser(
			ctx,
			targetBrowser,
			otherBrowsers,
			startAt,
			endAt,
			pageSize,
			nil,
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedResult := &LaggingFeatureCountPage{
			NextPageToken: nil,
			Metrics: []LaggingFeatureCount{
				{
					ReleaseDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					Count:       0,
				},
				{
					ReleaseDate: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
					Count:       0,
				},
				{
					ReleaseDate: time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
					Count:       1,
				},
				{
					ReleaseDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					Count:       0,
				},
				{
					ReleaseDate: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
					Count:       1,
				},
				{
					ReleaseDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
					Count:       1,
				},
				{
					ReleaseDate: time.Date(2024, 3, 28, 0, 0, 0, 0, time.UTC),
					Count:       2,
				},
			},
		}

		if !reflect.DeepEqual(expectedResult, result) {
			t.Errorf("unexpected result.\nExpected %+v\nReceived %+v", expectedResult, result)
		}
	})
}
