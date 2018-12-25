package rpc

import (
	"context"
	"fmt"

	"github.com/microsvs/base/pkg/errors"
	"github.com/microsvs/base/pkg/types"
	"github.com/microsvs/base/pkg/utils"
)

var (
	TOKEN_QUERY_SCHMEA = `
        query{
        	token(token: "%s") {
				token
				user_id
				token_expire
				refresh_token
				refresh_token_expire
        	}
        }
   `
	USER_QUERY_SCHEMA = `
		query{
			user(user_id: "%s") {
				id
				mobile
				nickname
			}
		}
   `
)

func GetUserIdFromTokenRPC(ctx context.Context, dns string, token string) (*types.Token, error) {
	var (
		data     map[string]interface{}
		err      error
		retToken = new(types.Token)
	)
	if data, err = CallService(ctx, dns, fmt.Sprintf(TOKEN_QUERY_SCHMEA, token)); err != nil {
		return nil, err
	}
	if err = utils.Decode(data, "token", retToken); err != nil {
		return nil, err
	}
	if len(retToken.UserId) <= 0 {
		return nil, errors.FGEInvalidToken
	}
	return retToken, nil
}

func GetUserFromIdRPC(ctx context.Context, dns string, id string) (*types.User, error) {
	var (
		data map[string]interface{}
		err  error
		user = new(types.User)
	)
	if data, err = CallService(ctx, dns, fmt.Sprintf(USER_QUERY_SCHEMA, id)); err != nil {
		return nil, err
	}
	if err = utils.Decode(data, "user", user); err != nil {
		return nil, err
	}
	if len(user.ID) <= 0 {
		return nil, errors.FGEInvalidUserID
	}
	return user, nil
}
