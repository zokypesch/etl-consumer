package data

// Response for response
type Response struct {
	Payload Payload `json:"payload"`
	Schema  Schema  `json:"schema"`
}

// Payload payload from kafka
type Payload struct {
	Before map[string]interface{} `json:"before"`
	After  map[string]interface{} `json:"after"`
	Source Source                 `json:"source"`
	TsMs   float64                `json:"ts_ms"`
	Op     string                 `json:"op"`
}

// Source for ourcing
type Source struct {
	Version   string  `json:"version"`
	Connector string  `json:"connector"`
	Name      string  `json:"name"`
	TsMs      float64 `json:"ts_ms"`
	Snapshot  string  `json:"snapshot"`
	Db        string  `json:"db"`
	Table     string  `json:"table"`
	ServerID  float64 `json:"server_id"`
	File      string  `json:"file"`
	Pos       float64 `json:"pos"`
	Row       float64 `json:"row"`
	Query     string  `json:"query"`
}

// Schema for schema
type Schema struct {
	Fields []Field `json:"fields"`
}

// Field for field
type Field struct {
	Type       string     `json:"type"`
	Optional   bool       `json:"optional"`
	Name       string     `json:"name"`
	Field      string     `json:"field"`
	Fields     []Field    `json:"fields"`
	Parameters Parameters `json:"parameters"`
}

// Parameters for parameters
type Parameters struct {
	Scale            string `json:"scale"`
	DecimalPrecision string `json:"connect.decimal.precision"`
}

// SearchFieldByName for search field by name
func SearchFieldByName(fields []Field, name string) Field {
	for _, v := range fields {
		if v.Field == name {
			return v
		}
	}
	return Field{}
}

// SearchFieldsByName for search fields
func (field *Field) SearchFieldsByName(name string) Field {
	for _, v := range field.Fields {
		if v.Field == name {
			return v
		}
	}
	return Field{}
}

/*{
	"payload": {
	  "before": null,
	  "after": {
		"id": 50,
		"province_id": 999,
		"seq": 0,
		"created_at": "2020-04-18T08:54:10Z"
	  },
	  "source": {
		"version": "1.1.1.Final",
		"connector": "mysql",
		"name": "dbserver2",
		"ts_ms": 0,
		"snapshot": "true",
		"db": "batch",
		"table": "batch_seq",
		"server_id": 0,
		"gtid": null,
		"file": "mysql-bin.000082",
		"pos": 289021,
		"row": 0,
		"thread": null,
		"query": null
	  },
	  "op": "c",
	  "ts_ms": 1587202401764,
	  "transaction": null
	}
  }*/
