package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)
// Переменные для хранения результатов и коэффициентов
var (
	koefficent_uspeha     = 0
	result            int = -2
	a                 int = -999999999
	last_text         string
)
// Функция для сравнения длины двух строк и возвращения короткой и длинной
func max_len(your_code, BDcode string) (string, string) {
	if len(your_code) < len(BDcode) {
		return your_code, BDcode
	}
	return BDcode, your_code
}

// Функция для расчета N-грамм между двумя строками
func N_gramms(first_code, second_code string) int {
	fmt.Println()
	fmt.Println("N_gramms")
	var res, interim_result int
	var max_volume float64
	first_code, second_code = max_len(first_code, second_code)
	if float64(len(first_code)) < 4 || float64(len(second_code)) < 4 {
		fmt.Println("NO SOLVE")
		return -1
	}
	if (float64(len(first_code)) / float64(len(second_code))) <= 0.6 {
		return -1
		fmt.Println("NO SOLVE")
	}
	if (float64(len(second_code)) / float64(len(first_code))) <= 0.6 {
		return -1
		fmt.Println("NO SOLVE")
	}
	first := []rune(first_code)
	max_volume = float64(len(second_code)) - 2
	for i := 2; i < len(first); i++ {
		var res_str string
		for j := i - 2; j <= i; j++ {
			res_str += string(first[j])
		}
		fmt.Print(res_str, "        ")
		if strings.Contains(second_code, res_str) {
			interim_result++
		}
	}
	fmt.Println()
	fmt.Println("max: ", max_volume, "  interim: ", interim_result)
	res = int((float64(interim_result) * 100 / max_volume))
	fmt.Println("lens: ", len(first_code), " ", len(second_code))
	fmt.Println(first_code, "     ", second_code)
	fmt.Println("N_GRAMMS: ", res)
	return res
}
// Функция для изменения флага
func flagChange(flag bool) bool {
	return !flag
}
// Функция для удаления символов из строки
func removeSimbols(code string) string {
	var res string
	correctSymbols := "+=-/&|!<>(){}*`[]%"
	r := []rune(code)
	flag1 := true
	flag2 := true
	for _, val := range r {
		symbol := string(val)
		if symbol == "'" {
			flag1 = flagChange(flag1)
		}
		if symbol == "\"" {
			flag2 = flagChange(flag2)
		}
		if flag1 && flag2 && strings.ContainsAny(correctSymbols, symbol) {
			res += symbol
		}
	}
	return res

}
// Функция для чтения программ из базы данных
func readProgramms(db *sql.DB, code string) int {
	rows, err := db.Query("SELECT Code FROM Programms")
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer rows.Close()
	max := 0
	for rows.Next() {
		var Code string
		err := rows.Scan(&Code)
		if err != nil {
			fmt.Println("STROKI ERROR")
			return -1
		}
		//fmt.Println("SCAN BD: ", Code)
		variable := N_gramms(Code, code)
		if max < variable {
			max = variable
			fmt.Println("max size: ", variable)
		}
		fmt.Println()
	}
	fmt.Println("Read")
	return max
}
// Функция для добавления программы в базу данных
func addProgramm(db *sql.DB, code string) {
	if len(removeSimbols(code)) < 4 {
		return
	}
	name := code
	stmt, err := db.Prepare("INSERT INTO Programms (Code) VALUES (?)")
	if err != nil {
		fmt.Println("ADD ERROR 1")
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(name)
	if err != nil {
		fmt.Println("ADD ERROR 2")
		return
	}
	fmt.Println("ADD successful")
}
// Функция для работы с базой данных
func DBfunc(code string) int {
	db, err := sql.Open("sqlite3", "./Programms.db")
	if err != nil {
		fmt.Printf("ОШИБКА В БАЗЕ ДАННЫХ С ОТКРЫТИЕМ: %v\n", err)
		//http.Error("Ошибка при работе с базой данных", http.StatusInternalServerError)
	}
	defer db.Close()
	result = readProgramms(db, code)
	if result <= 70 {
		addProgramm(db, code)
		fmt.Println("WAS ADD, result: ", result)
	}
	fmt.Println("ANSWER: ", result)
	return result
}
// Функция для обработки запросов на главную страницу
func home_page(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("home_page.html")
	tmpl.Execute(w, nil)
}
// Функция для обработки запросов на отправку кода
func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		text := r.FormValue("code")
		if text == "" {
			http.Error(w, "Поле текста является обязательным", http.StatusBadRequest)
			return
		}

		if text != last_text {
			result = DBfunc(removeSimbols(text))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"result": %d}`, result-koefficent_uspeha)
		last_text = text
	} else {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
	}

}
// Функция для регистрации обработчиков запросов
func HandleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/send", handleSend)
	http.ListenAndServe(":8080", nil)
}

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	HandleRequest()
}
