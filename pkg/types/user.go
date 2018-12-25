package types

import (
	"time"

	"github.com/graphql-go/graphql"
)

type STATUS int16

const (
	// 记录状态:  10：有效；20：无效
	STATUS__OK      STATUS = 10
	STATUS__INVALID STATUS = 20
)

type User struct {
	ID        string    `mapstructure:"user_id" msgpack:"user_id" db:"user_id" json:"id"`
	Mobile    string    `msgpack:"mobile" db:"mobile" json:"mobile"`
	Name      string    `mapstructure:"nickname" msgpack:"nickname" db:"nickname" json:"nickname"`
	Status    int       `msgpack:"status" db:"status" json:"status"` // -10: 无效；10: 有效
	UpdatedAt time.Time `msgpack:"updated_at" db:"updated_at" json:"updated_at"`
	CreatedAt time.Time `msgpack:"created_at" db:"created_at" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

var GLUser = graphql.NewObject(GLUserConfig)
var GLUserConfig = graphql.ObjectConfig{
	Name: "BasicUser",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.String,
			Description: "用户ID",
		},
		"nickname": &graphql.Field{
			Type:        graphql.String,
			Description: "用户昵称",
		},
		"mobile": &graphql.Field{
			Type:        graphql.String,
			Description: "电话号码，数据库保存加密以后的电话号码，返回脱敏以后的电话号码",
		},
		"status": &graphql.Field{
			Type:        GLUserStatus,
			Description: "用户状态",
		},
	},
}
var GLUserStatus = graphql.NewEnum(graphql.EnumConfig{
	Name:        "BasicUserStatus",
	Description: "用户状态",
	Values: graphql.EnumValueConfigMap{
		"Normal": &graphql.EnumValueConfig{
			Value:       STATUS__OK,
			Description: "正常状态",
		},
		"Delete": &graphql.EnumValueConfig{
			Value:       STATUS__INVALID,
			Description: "被删除状态",
		},
	},
})
