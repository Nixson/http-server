package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var jwtTokenOpenKey *string = nil

type openTokenResponse struct {
	Alg   *string `json:"alg"`
	Value *string `json:"value"`
}

func JWTTokenOpenKey() (*string, error) {
	if params.Env.GetBool("security.enable") {
		if jwtTokenOpenKey == nil {
			tokenKeyUrl := params.Env.GetString("security.token")

			client := &http.Client{}
			req, _ := http.NewRequest("GET", tokenKeyUrl, nil)
			req.Header.Add("Content-Type", "application/json")
			tokenKeyResponse, err := client.Do(req)
			if err != nil {
				return nil, err
			}

			tokenKeyBytes, err := ioutil.ReadAll(tokenKeyResponse.Body)
			if err != nil {
				return nil, err
			}

			var openToken = new(openTokenResponse)
			err = json.Unmarshal(tokenKeyBytes, openToken)
			if err != nil {
				return nil, err
			}

			tokenKey := openToken.Value
			jwtTokenOpenKey = tokenKey
			return jwtTokenOpenKey, nil
		} else {
			return jwtTokenOpenKey, nil
		}
	} else {
		return nil, errors.New("jwt security is disabled")
	}
}
