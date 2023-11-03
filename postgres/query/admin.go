package query

func insertAdmin() string {
	return `
		INSERT INTO users(
			name,
			lastname,
			email,
			password,
			active
		) 	VALUES ($1,$2,$3,$4,$5,$6)`
}
