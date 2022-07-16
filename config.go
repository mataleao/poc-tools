package poctools

type DatabaseSettings struct {
	DbClusterIdentifier string `json:"dbClusterIdentifier"`
	Password            string `json:"password"`
	Engine              string `json:"engine"`
	Port                int    `json:"port"`
	Host                string `json:"host"`
	Username            string `json:"username"`
	DatabaseName        string
}

var DefaultPaginationLimit int64
