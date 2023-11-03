package query


func insertUser() string {
	return `
		INSERT INTO users(
			name,
			lastname,
			email,
			password,
			phone
		) 	VALUES ($1,$2,$3,$4,$5,$6)`
}
