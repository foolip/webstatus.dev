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

import { type CSSResultGroup, LitElement, type TemplateResult, html } from 'lit'
import { customElement, property } from 'lit/decorators.js'

import { type AppSettings } from '../../../common/app-settings.js'
import { SHARED_STYLES } from '../css/shared-css.js'
import './webstatus-app-settings.js'
import './webstatus-header.js'
import './webstatus-page.js'

@customElement('webstatus-app')
export class WebstatusApp extends LitElement {
  @property({ type: Object })
    settings!: AppSettings

  static get styles (): CSSResultGroup {
    return [
      SHARED_STYLES
    ]
  }

  protected render (): TemplateResult {
    return html`
      <webstatus-app-settings .appSettings='${this.settings}'
        >
        <webstatus-header></webstatus-header>
        <webstatus-page>
          <slot></slot>
        </webstatus-page>
      </webstatus-app-settings>
    `
  }
}
