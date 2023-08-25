-- +goose Up
CREATE TABLE Todos (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);


-- +goose Down
DROP TABLE Todos;
