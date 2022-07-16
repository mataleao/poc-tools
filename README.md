

poctools.SetDbEngine(databaseEngine)




sqlExec := poctools.CreateSqlExecutor(R.db)
sql := "your query"
poctools.FindAllPaged(s, sql, params, R.getLeadsDto)

