package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBname   string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBname, cfg.SSLMode)
}

func main() {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "pgeazy",
		Password: "pgeazypassword",
		DBname:   "eazyweather",
		SSLMode:  "disable",
	}
	db, err := sql.Open("pgx", cfg.String())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("db is connected")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT,
			email TEXT UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			amount INT,
			description TEXT
		);
	`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Tables created")

	// id := 1
	// var name, email string
	// row := db.QueryRow(`
	// 	SELECT name,email
	// 	FROM users
	// 	WHERE id=$1;`, id)
	// err = row.Scan(&name, &email)
	// if err == sql.ErrNoRows {
	// 	fmt.Println("error is for no rows!")
	// } else if err != nil {
	// 	fmt.Println("i can panic here")
	// } else {
	// 	fmt.Printf("Users returned. name = %s and email = %s \n", name, email)
	// }
	user_id := 1
	// for i := 1; i <= 5; i++ {
	// 	amount := i * 100
	// 	description := fmt.Sprintf("sometihng i wanna sell you is like this #%d", i)
	// 	db.Exec(`
	// 		INSERT INTO orders(user_id, amount, description)
	// 		VALUES ($1,$2,$3)`, user_id, amount, description)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	// fmt.Println("Created orders")

	rows, err := db.Query(`
		SELECT id,amount,description
		FROM orders
		WHERE user_id=$1`, user_id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	type Order struct {
		Id          int
		UserID      int
		Amount      int
		Description string
	}
	var orders []Order

	for rows.Next() {
		var order Order
		order.UserID = user_id
		err := rows.Scan(&order.Id, &order.Amount, &order.Description)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	fmt.Println("Orders: ", orders)
	// name := "sdsd singh"
	// email := "dsds@live.com"
	//query := fmt.Sprintf(`
	//	INSERT INTO users(name, email)
	//	VALUES ('%s','%s');`, name, email)
	//fmt.Println("executing quiery: ", query)
	//_, err = db.Exec(query)

	// row := db.QueryRow(`
	// 	INSERT INTO users(name, email)
	// 	VALUES ($1,$2) RETURNING id;`, name, email)
	// var id int
	// err = row.Scan(&id)  xyti6776u zcc cfgwqcf xxmc   mn  nphgs    CNYR4gpjkf2345yuikozzccc vRvvvbhhh
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Users created. id = ", id)

}
