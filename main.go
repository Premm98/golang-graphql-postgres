package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
	handler "github.com/graphql-go/graphql-go-handler"
	_ "github.com/lib/pq"
)

type Employee struct {
	ID          int
	Name, Email string
	Password    string
}

var employees []Employee

const (
	hostname      = "localhost"
	host_port     = 5432
	username      = "postgres"
	password      = "123456"
	database_name = "Kibbcom"
)

var employeeType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Employee",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"password": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query1",
	Fields: graphql.Fields{
		"names": &graphql.Field{
			Type:        graphql.NewList(employeeType),
			Description: "All Jobs",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return employees, nil
			},
		},

		"name": &graphql.Field{
			Type: employeeType,
			Args: graphql.FieldConfigArgument{
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id := params.Args["email"].(string)
				for _, emp := range employees {
					if emp.Email == id {
						return emp, nil
					}
				}
				return nil, nil
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootMutation",
	Fields: graphql.Fields{
		"createUser": &graphql.Field{
			Type: employeeType,
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},

				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},

				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},

			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				name := params.Args["name"].(string)
				email := params.Args["email"].(string)
				pass := params.Args["password"].(string)

				newEmployee := &Employee{
					Name:     name,
					Email:    email,
					Password: pass,
				}

				connStr := fmt.Sprintf("port=%d host=%s user=%s password=%s dbname=%s sslmode=disable", host_port, hostname, username, password, database_name)
				db, err := sql.Open("postgres", connStr)
				if err != nil {
					fmt.Println(`Could not connect to db`)
					panic(err)
				}
				defer db.Close()

				insertStatement := `
					INSERT INTO employee (name, email, password)
					VALUES ($1, $2, $3)`

				_, err = db.Query(insertStatement, name, email, pass)

				return newEmployee, nil

			},
		},
	},
})

func main() {
	connStr := "postgres://postgres:123456@localhost/Kibbcom?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(`Could not connect to db`)
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM employee`)
	if err != nil {
		panic(err)
	}

	var name, email string
	var password string
	var ID int

	for rows.Next() {
		rows.Scan(&ID, &name, &email, &password)
		row := Employee{ID, name, email, password}
		employees = append(employees, row)
	}

	var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})

	h := handler.New(&handler.Config{
		Schema:   &Schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)

	fmt.Println("Server Started...lalalalalala")

	http.ListenAndServe(":8080", nil)
}
