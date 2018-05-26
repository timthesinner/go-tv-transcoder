//By TimTheSinner
package main

import (
	"regexp"
)

/**
 * Copyright (c) 2016 TimTheSinner All Rights Reserved.
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

func groupsFromRegex(r *regexp.Regexp, search string) (paramsMap map[string]string) {
	match := r.FindStringSubmatch(search)
	if len(match) == 0 {
		return
	}

	paramsMap = make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			paramsMap[name] = match[i]
		}
	}
	return
}
