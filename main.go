package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Db *sql.DB
var err error

type User struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ファイル出力を受け取るための構造体
type Logdata struct {
	User  User   `json:"user"`
	Dist  string `json:"dist"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Src   string `json:"src"`
	Time  string `json:"time"`
}

func main() {
	Db, err = sql.Open("postgres", "user=user dbname=gotask password=password sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	defer Db.Close()

	cmd := `create table if not exists users(
		id serial primary key,
		age integer,
		name varchar(500),
		role char(15)
	)`
	_, err := Db.Exec(cmd)
	if err != nil {
		log.Fatalln(err)
	}

	// 指定するファイル名が複数ある場合にはエラーを返す
	args := os.Args
	if len(args) != 2 {
		log.Fatalln("指定可能なファイルは一つです。")
	}

	file, err := os.Open(args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// ファイルを一行ずつ読み込む
	scanner := bufio.NewScanner(file)
	transaction, err := Db.Begin()
	defer func() {
		if recover() != nil {
			transaction.Rollback()
		}
	}()

	cmd = `INSERT INTO users (age, name, role) VALUES ($1, $2, $3)`

	for scanner.Scan() {
		var jsonByte []byte
		jsonByte = append(jsonByte, scanner.Text()...)
		var logData Logdata
		// データをJSONに変換
		err := json.Unmarshal(jsonByte, &logData)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = transaction.Exec(
			cmd,
			logData.User.Age,
			logData.User.Name,
			logData.User.Role,
		)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err != nil {
		transaction.Rollback()
	} else {
		transaction.Commit()
	}
}
