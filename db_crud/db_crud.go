package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	DB *sql.DB
}

var (
	defaultOffset = 0
	defaultLimit  = 5
)

func NewDbCRUD(db *sql.DB) (http.Handler, error) {
	// ваша реализация тут
	handler := &Handler{
		DB: db,
	}

	siteMux := http.NewServeMux()
	siteMux.HandleFunc("/", handler.ServeHTTP)

	return siteMux, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Path

	params := strings.Split(query, "/")
	var response interface{}
	var handlerError error

	if len(params) == 2 && params[1] == "" {
		response, handlerError = h.GetTableNames()
	} else if len(params) == 2 || len(params) == 3 && params[2] == "" {
		tableName := params[1]

		if r.Method == http.MethodGet {
			offsetStr := r.FormValue("offset")
			offset, err := strconv.Atoi(offsetStr)
			if err != nil || (offset == 0) {
				offset = defaultOffset
			}

			limitStr := r.FormValue("limit")
			limit, err := strconv.Atoi(limitStr)
			if err != nil || (limit == 0) {
				limit = defaultLimit
			}
			response, handlerError = h.GetFromTable(tableName, offset, limit)
		} else if r.Method == http.MethodPut {
			var newItem map[string]interface{}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(body, &newItem)
			if err != nil {
				panic(err)
			}

			response, handlerError = h.InsertIntoTable(tableName, newItem)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if len(params) == 3 {
		tableName := params[1]
		itemID := params[2]
		if r.Method == http.MethodGet {
			response, handlerError = h.GetFromTableByID(tableName, itemID)
		} else if r.Method == http.MethodPost {
			var newItem map[string]interface{}

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(body, &newItem)
			if err != nil {
				panic(err)
			}

			response, handlerError = h.UpdateByID(tableName, itemID, newItem)
		} else if r.Method == http.MethodDelete {
			response, handlerError = h.DeleteByID(tableName, itemID)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}

	finalResponse := make(map[string]interface{})

	if handlerError != nil {
		switch handlerError.Error() {
		case "Not found":
			w.WriteHeader(http.StatusNotFound)
		case "Bad request":
			w.WriteHeader(http.StatusBadRequest)
		}
		finalResponse["error"] = response
	} else {
		finalResponse["response"] = response
	}

	finalResponseJSON, err := json.Marshal(finalResponse)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, string(finalResponseJSON))

}

func (h *Handler) GetTableNames() (interface{}, error) {
	tables, err := h.DB.Query("SHOW TABLES")
	if err != nil {
		panic(err)
	}

	res := map[string][]string{}

	var tableName string
	for tables.Next() {
		err = tables.Scan(&tableName)
		if err != nil {
			panic(err)
		}
		res["tables"] = append(res["tables"], tableName)
	}
	tables.Close()

	return res, nil
}

func (h *Handler) GetFromTable(tableName string, offset int, limit int) (interface{}, error) {
	resp, err := h.HasTableName(tableName)
	if err != nil {
		return resp, err
	}

	items, err := h.DB.Query("SELECT * FROM "+tableName+" LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		panic(err)
	}
	result := []map[string]interface{}{}

	columns, err := items.Columns()
	if err != nil {
		panic(err)
	}
	columnTypes, err := items.ColumnTypes()
	if err != nil {
		panic(err)
	}

	for items.Next() {
		itemStruct := map[string]interface{}{}

		rawResult := make([][]byte, len(columns))
		item := make([]interface{}, len(columns))
		for i, _ := range rawResult {
			item[i] = &rawResult[i]
		}

		err = items.Scan(item...)
		if err != nil {
			panic(err)
		}

		for i, value := range rawResult {
			if nullable, ok := columnTypes[i].Nullable(); len(value) == 0 && ok && nullable {
				itemStruct[columns[i]] = nil
			} else if strings.HasPrefix(columnTypes[i].ScanType().Name(), "int") {
				valueInt, err := strconv.Atoi(string(value))
				if err != nil {
					panic(err)
				}
				itemStruct[columns[i]] = valueInt
			} else {
				itemStruct[columns[i]] = string(value)
			}
		}
		result = append(result, itemStruct)
	}
	items.Close()

	records := make(map[string]interface{})
	records["records"] = result

	return records, nil
}

func (h *Handler) GetColumnsFromTable(tableName string) ([]string, []map[string]string) {
	columns, err := h.DB.Query("SHOW FULL COLUMNS FROM " + tableName)
	if err != nil {
		panic(err)
	}

	rawResult := make([][]byte, 9)
	result := []string{}
	column := make([]interface{}, 9)
	for i, _ := range rawResult {
		column[i] = &rawResult[i]
	}

	info := []map[string]string{}
	for columns.Next() {
		err = columns.Scan(column...)
		if err != nil {
			panic(err)
		}

		extra := map[string]string{"extra": string(rawResult[6])}
		extra["key"] = string(rawResult[4])
		extra["null"] = string(rawResult[3])
		if strings.HasPrefix(string(rawResult[1]), "varchar") || strings.HasPrefix(string(rawResult[1]), "text") {
			extra["type"] = "string"
		} else if strings.HasPrefix(string(rawResult[1]), "int") {
			extra["type"] = "int"
		} else if strings.HasPrefix(string(rawResult[1]), "float") {
			extra["type"] = "float"
		}

		info = append(info, extra)
		result = append(result, string(rawResult[0]))
	}
	columns.Close()

	return result, info
}

func (h *Handler) InsertIntoTable(tableName string, newItem map[string]interface{}) (interface{}, error) {
	resp, err := h.HasTableName(tableName)
	if err != nil {
		return resp, err
	}

	columns, extra := h.GetColumnsFromTable(tableName)
	insertColumns := []string{}
	questionMarks := []string{}

	params := []interface{}{}
	for i, column := range columns {
		if extra[i]["extra"] != "auto_increment" {
			insertColumns = append(insertColumns, column)
			questionMarks = append(questionMarks, "?")
			params = append(params, newItem[column])
		}
	}

	result, err := h.DB.Exec("INSERT INTO "+tableName+" ("+strings.Join(insertColumns, ",")+") VALUES ("+strings.Join(questionMarks, ",")+")", params...)
	if err != nil {
		panic(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{"id": id}, nil
}

func (h *Handler) HasTableName(tableName string) (interface{}, error) {
	tablesInterface, err := h.GetTableNames()
	if err != nil {
		panic(err)
	}
	tables := tablesInterface.(map[string][]string)

	existTable := false
	for _, table := range tables["tables"] {
		if table == tableName {
			existTable = true
		}
	}

	if !existTable {
		return "unknown table", fmt.Errorf("Not found")
	} else {
		return nil, nil
	}
}

func (h *Handler) GetFromTableByID(tableName string, itemID string) (interface{}, error) {
	resp, err := h.HasTableName(tableName)
	if err != nil {
		return resp, err
	}

	items, err := h.DB.Query("SELECT * FROM "+tableName+" WHERE id = ?", itemID)
	if err != nil {
		panic(err)
	}
	result := []map[string]interface{}{}

	columns, err := items.Columns()
	if err != nil {
		panic(err)
	}
	columnTypes, err := items.ColumnTypes()
	if err != nil {
		panic(err)
	}

	for items.Next() {
		itemStruct := map[string]interface{}{}

		rawResult := make([][]byte, len(columns))
		item := make([]interface{}, len(columns))
		for i, _ := range rawResult {
			item[i] = &rawResult[i]
		}

		err = items.Scan(item...)
		if err != nil {
			panic(err)
		}

		for i, value := range rawResult {
			if nullable, ok := columnTypes[i].Nullable(); len(value) == 0 && ok && nullable {
				itemStruct[columns[i]] = nil
			} else if strings.HasPrefix(columnTypes[i].ScanType().Name(), "int") {
				valueInt, err := strconv.Atoi(string(value))
				if err != nil {
					panic(err)
				}
				itemStruct[columns[i]] = valueInt
			} else {
				itemStruct[columns[i]] = string(value)
			}
		}
		result = append(result, itemStruct)
	}
	items.Close()

	records := make(map[string]interface{})
	if len(result) == 1 {
		records["record"] = result[0]
	} else if len(result) > 0 {
		records["records"] = result
	} else {
		return "record not found", fmt.Errorf("Not found")
	}

	return records, nil
}

func (h *Handler) UpdateByID(tableName string, itemID string, newItem map[string]interface{}) (interface{}, error) {
	resp, err := h.HasTableName(tableName)
	if err != nil {
		return resp, err
	}

	columns, extra := h.GetColumnsFromTable(tableName)

	insertColumns := []string{}
	insertValues := []interface{}{}

	for k, v := range newItem {
		existColumn := false
		for i, column := range columns {
			if column == k {
				existColumn = true

				ok := false
				if newItem[k] == nil && extra[i]["null"] == "YES" {
					ok = true
				} else if extra[i]["type"] == "int" {
					_, ok = newItem[k].(int)
				} else if extra[i]["type"] == "string" {
					_, ok = newItem[k].(string)
				} else if extra[i]["type"] == "float" {
					_, ok = newItem[k].(float64)
				}

				if !ok || extra[i]["key"] == "PRI" {
					return "field " + column + " have invalid type", fmt.Errorf("Bad request")
				}
			}
		}

		if existColumn {
			insertColumns = append(insertColumns, k)
			insertValues = append(insertValues, v)
		}
	}
	insertValues = append(insertValues, itemID)

	result, err := h.DB.Exec("UPDATE "+tableName+" SET "+strings.Join(insertColumns, " = ?, ")+" = ? WHERE id = ?", insertValues...)
	if err != nil {
		panic(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	return map[string]interface{}{"updated": rowsAffected}, nil
}

func (h *Handler) DeleteByID(tableName string, itemID string) (interface{}, error) {
	resp, err := h.HasTableName(tableName)
	if err != nil {
		return resp, err
	}

	result, err := h.DB.Exec("DELETE FROM "+tableName+" WHERE id = ?", itemID)
	if err != nil {
		panic(err)
	}

	rowsAffected, err := result.RowsAffected()
	return map[string]interface{}{"deleted": rowsAffected}, nil
}
