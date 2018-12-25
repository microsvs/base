package types

import (
	"time"

	"github.com/graphql-go/graphql"
)

type Token struct {
	UserId             string    `db:"user_id" json:"user_id"`
	Token              string    `db:"token" json:"token"`
	TokenExpire        time.Time `db:"token_expire" json:"token_expire"`
	RefreshToken       string    `db:"refresh_token" "json:"refresh_token"`
	RefreshTokenExpire time.Time `db:"refresh_token_expire" "json:"refresh_token_expire"`
}

func (Token) TableName() string {
	return "tokens"
}

var GLToken = graphql.NewObject(GLTokenConfig)

//GLTokenConfig GraphQL Token配置
var GLTokenConfig = graphql.ObjectConfig{
	Name: "Token",
	Fields: graphql.Fields{
		"user_id": &graphql.Field{
			Type:        graphql.String,
			Description: "Token对应的用户ID",
		},
		"token": &graphql.Field{
			Type:        graphql.String,
			Description: "Token",
		},
		"token_expire": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Token过期时间",
		},
		"refresh_token": &graphql.Field{
			Type:        graphql.String,
			Description: "使用RefreshToken进行刷新操作",
		},
		"refresh_token_expire": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "RefreshToken的过期时间，如果RefreshToken已过期则需要重新登录",
		},
	},
}
