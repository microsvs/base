package base

import (
	"fmt"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
)

type Person struct {
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Vehicle   Vehicle   `json:"vehicle"`
	Workers   []Worker  `json:"workers"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Worker struct {
	Company  string   `json:"company"`
	Position Position `json:"position"`
}

type Position int16

var (
	RD  Position = 10
	PD  Position = 20
	CEO Position = 30
	CTO Position = 40
)

type Vehicle int16

var (
	Bike       Vehicle = 10
	Motorcycle Vehicle = 20
	Car        Vehicle = 30
)

var GLPerson = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "Person",
		Description: "个人信息",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "姓名",
			},
			"age": &graphql.Field{
				Type:        graphql.Int,
				Description: "年龄",
			},
			"vehicle": &graphql.Field{
				Type:        GLVehicleEnum,
				Description: "每日交通工具",
			},
			"workers": &graphql.Field{
				Type:        graphql.NewList(GLWorker),
				Description: "身份列表",
			},
			"updated_at": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "更新时间",
			},
			"created_at": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "创建时间",
			},
		},
	},
)

var GLWorker = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "Worker",
		Description: "身份",
		Fields: graphql.Fields{
			"company": &graphql.Field{
				Type:        graphql.String,
				Description: "公司名称",
			},
			"position": &graphql.Field{
				Type:        GLPositionEnum,
				Description: "职位",
			},
		},
	},
)
var GLPositionEnum = graphql.NewEnum(
	graphql.EnumConfig{
		Name:        "PositionEnum",
		Description: "职位类型",
		Values: graphql.EnumValueConfigMap{
			"RD": &graphql.EnumValueConfig{
				Value:       RD,
				Description: "研发",
			},
			"PD": &graphql.EnumValueConfig{
				Value:       PD,
				Description: "产品",
			},
			"CEO": &graphql.EnumValueConfig{
				Value:       CEO,
				Description: "公司职业经理人",
			},
			"CTO": &graphql.EnumValueConfig{
				Value:       CTO,
				Description: "首席技术官",
			},
		},
	},
)
var GLVehicleEnum = graphql.NewEnum(
	graphql.EnumConfig{
		Name:        "VehicleEnum",
		Description: "交通工具方式",
		Values: graphql.EnumValueConfigMap{
			"Bike": &graphql.EnumValueConfig{
				Value:       Bike,
				Description: "自行车",
			},
			"Motorcycle": &graphql.EnumValueConfig{
				Value:       Motorcycle,
				Description: "摩托车",
			},
			"Car": &graphql.EnumValueConfig{
				Value:       Car,
				Description: "汽车",
			},
		},
	},
)

// simple test: for enum data convert
func TestFixTypeFromGraphqlToGoSimple(t *testing.T) {
	fmt.Println(FixTypeFromGraphqlToGo("Bike", GLVehicleEnum))
	fmt.Println(FixTypeFromGraphqlToGo("Motorcycle", GLVehicleEnum))
	fmt.Println(FixTypeFromGraphqlToGo("Car", GLVehicleEnum))
}

// complex test: for struct data convert
func TestFixTypeFromGraphqlToGoComplext(t *testing.T) {
	var person = map[string]interface{}{
		"name":    "kim",
		"age":     16,
		"vehicle": "Bike",
		"workers": []interface{}{
			map[string]interface{}{
				"company":  "alibaba",
				"position": "CEO",
			},
			map[string]interface{}{
				"company":  "tencent",
				"position": "CTO",
			},
		},
		"updated_at": "2019-11-24T03:29:08+08:00",
		"created_at": "2019-11-24T03:29:08+08:00",
	}
	fmt.Println(FixTypeFromGraphqlToGo(person, GLPerson))
}

func TestFixTypeFromGoToGraphqlSimple(t *testing.T) {
	fmt.Println(FixTypeFromGoToGraphql(CEO, GLPositionEnum))
}

// FixTypeFromGoToGraphql complex unit test
func TestFixTypeFromGoToGraphqlComplex(t *testing.T) {
	var person = &Person{
		Name:    "kim",
		Age:     16,
		Vehicle: Bike,
		Workers: []Worker{
			Worker{
				Company:  "alibaba",
				Position: CEO,
			},
			Worker{
				Company:  "tencent",
				Position: CTO,
			},
		},
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	fmt.Println(FixTypeFromGoToGraphql(person, GLPerson))
}
