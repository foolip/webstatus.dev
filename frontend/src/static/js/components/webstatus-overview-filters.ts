/**
 * Copyright 2023 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {LitElement, type TemplateResult, CSSResultGroup, css, html} from 'lit';
import {customElement} from 'lit/decorators.js';
import {getSearchQuery} from '../utils/urls.js';
import {SHARED_STYLES} from '../css/shared-css.js';
import {SlInput, SlMenu, SlMenuItem} from '@shoelace-style/shoelace';

import './webstatus-overview-table.js';

@customElement('webstatus-overview-filters')
export class WebstatusOverviewFilters extends LitElement {
  location!: {search: string}; // Set by parent.

  static get styles(): CSSResultGroup {
    return [
      SHARED_STYLES,
      css`
        .all-filter-controls,
        .filter-by-feature-name,
        .filter-buttons {
          gap: var(--content-padding);
        }
      `,
    ];
  }

  firstUpdated(): void {

    const makeFilterSelectHandler = (id: string): ((event: Event) => void) => {
      return (event: Event) => {
        const menu = event.target as SlMenu;
        const menuChildren = menu.children;

        const menuItemsArray: Array<SlMenuItem> = Array.from(
          menuChildren
        ).filter(child => child instanceof SlMenuItem) as Array<SlMenuItem>;

        // Create a list of the currently checked sl-menu-items.
        // Build a query string from the values of those items.
        const checkedItems = menuItemsArray.filter(
          menuItem => menuItem.checked
        );
        const checkedItemsValues = checkedItems.map(menuItem => menuItem.value);
        const filterQueryItems = checkedItemsValues.join(',');
        const filterQueryWithPrefix = `${id}:${filterQueryItems}`;

        // Get the filter-query-input value.
        const filterQueryInput = this.shadowRoot!.getElementById(
          'filter-query-input'
        ) as SlInput;

        // Modify the filterQueryInput to reflect the currently selected items.
        // For now, just overwrite the entire query string.
        filterQueryInput.value = filterQueryWithPrefix;
      };
    };

    // Add sl-select event handler to all sl-menu elements.
    const menuElements = Array.from(
      this.shadowRoot!.querySelectorAll('sl-menu')
    );
    for (const menuElement of menuElements) {
      const id = menuElement.id;
      menuElement.addEventListener('sl-select', makeFilterSelectHandler(id));
    }
  }

  render(): TemplateResult {
    const query = getSearchQuery(this.location);
    return html`
      <div class="vbox all-filter-controls">
        <div class="hbox filter-by-feature-name">
          <sl-input
            id="filter-query-input"
            class="halign-stretch"
            placeholder="Filter by feature name..."
            value="${query}"
          >
            <sl-icon name="search" slot="prefix"></sl-icon>
          </sl-input>
        </div>

        <div class="hbox wrap filter-buttons">
          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Available on
            </sl-button>
            <sl-menu id="available_on">
              <sl-menu-item type="checkbox" value="chrome">
                Chrome
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="edge"> Edge </sl-menu-item>
              <sl-menu-item type="checkbox" value="firefox">
                Firefox
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="safari">
                Safari
              </sl-menu-item>
            </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Not available on
            </sl-button>
            <sl-menu id="not_available_on">
              <sl-menu-item type="checkbox" value="chrome">
                Chrome
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="edge"> Edge </sl-menu-item>
              <sl-menu-item type="checkbox" value="firefox">
                Firefox
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="safari">
                Safari
              </sl-menu-item>
              <sl-divider></sl-divider>
              <sl-menu-item type="checkbox" value="not-in">
                Not available in 1 browser
              </sl-menu-item>
            </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Baseline status
            </sl-button>
            <sl-menu id="baseline_status">
              <sl-menu-item type="checkbox" value="widely">
                Widely available
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="newly">
                Newly available
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="limited">
                Limited availability
              </sl-menu-item>
            </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Browser type
            </sl-button>
            <sl-menu id="browser_type">
              <sl-menu-item type="checkbox" value="stable-builds">
                Stable builds
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="dev-builds">
                Dev builds
              </sl-menu-item>
            </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Standards track
            </sl-button>
            <sl-menu> </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Spec maturity
            </sl-button>
            <sl-menu id="spec_maturity">
              <sl-menu-item type="checkbox" value="unknown">
                Unknown
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="proposed">
                Proposed
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="incubation">
                Incubation
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="working-draft">
                Working draft
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="living-standard">
                Living standard
              </sl-menu-item>
            </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon slot="prefix" name="plus-circle"></sl-icon>
              Web platform test score
            </sl-button>
            <sl-menu id="web_platform_test_score">
              <sl-menu-item type="checkbox" value="chrome">
                Chrome
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="edge"> Edge </sl-menu-item>
              <sl-menu-item type="checkbox" value="firefox">
                Firefox
              </sl-menu-item>
              <sl-menu-item type="checkbox" value="safari">
                Safari
              </sl-menu-item>
            </sl-menu>
          </sl-dropdown>

          <sl-dropdown stay-open-on-select>
            <sl-button slot="trigger">
              <sl-icon
                slot="prefix"
                name="square-split-horizontal"
                library="phosphor"
              ></sl-icon>
              Columns
            </sl-button>
          </sl-dropdown>
        </div>
      </div>
    `;
  }
}
