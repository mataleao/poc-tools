

poctools.SetDbEngine(databaseEngine)




sqlExec := poctools.CreateSqlExecutor(R.db)
sql := "your query"
return poctools.FindAllPaged(sql, sqlExec, params, "entity name", R.getLeadsDto)
