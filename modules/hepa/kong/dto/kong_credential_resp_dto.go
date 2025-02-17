// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dto

type KongCredentialListDto struct {
	Total int64               `json:"total"`
	Data  []KongCredentialDto `json:"data"`
}

type KongCredentialDto struct {
	ConsumerId string `json:"consumer_id,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	Id         string `json:"id,omitempty"`
	// key-auth, sign-auth
	Key string `json:"key,omitempty"`
	// oauth2
	RedirectUrl interface{} `json:"redirect_uri,omitempty"`
	// v2
	RedirectUrls []string `json:"redirect_uris,omitempty"`
	Name         string   `json:"name,omitempty"`
	ClientId     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	// sign-auth, hmac-auth
	Secret string `json:"secret,omitempty"`
	// hmac-auth
	Username string `json:"username,omitempty"`
}

func (dto *KongCredentialDto) ToHmacReq() {
	if dto == nil {
		return
	}
	if dto.Username == "" {
		dto.Username = dto.Key
		dto.Key = ""
	}
}

func (dto *KongCredentialDto) ToHmacResp() {
	if dto == nil {
		return
	}
	if dto.Key == "" {
		dto.Key = dto.Username
		dto.Username = ""
	}
}

func (dto *KongCredentialDto) ToV2() {
	if dto == nil {
		return
	}
	if url, ok := dto.RedirectUrl.([]string); ok && len(url) > 0 {
		dto.RedirectUrls = url
		dto.RedirectUrl = nil
	}
}

func (dto *KongCredentialDto) Compatiable() {
	if dto == nil {
		return
	}
	dto.RedirectUrl = dto.RedirectUrls
}
