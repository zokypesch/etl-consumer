package scheme

// Response for response
type Response struct {
	Payload Payload `json:"payload"`
}

// Payload payload from kafka
type Payload struct {
	Source       Source `json:"source"`
	DatabaseName string `json:"databaseName"`
	DDL          string `json:"ddl"`
}

// Source for table source
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
