package database

const schemaSQL = `
CREATE TABLE IF NOT EXISTS tweets(
    tweet_id VARCHAR(255) unique ,
    time TIMESTAMP
);
`

const InsertionQuery = "INSERT INTO tweets (tweet_id, time) VALUES ($1, $2)"
const DeleteQuery = "delete from tweets where time <$1"
