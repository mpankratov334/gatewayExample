package repo

const (
	InitQuery = `CREATE TABLE IF NOT EXISTS users (
						id SERIAL PRIMARY KEY, -- Уникальный идентификатор пользователя, UUID хорош для распределённых систем.
						email VARCHAR(255) UNIQUE, -- Основной идентификатор для входа и связи с пользователем. Ограничение уникально
						username VARCHAR(50) UNIQUE NOT NULL, -- Уникальное публичное имя пользователя, может использоваться для отображения или
						password_hash TEXT NOT NULL, -- Хранит хэш пароля пользователя. Никогда не храним пароли в открытом виде. !!! Д
						first_name VARCHAR(100), -- Имя пользователя, для отображения или персонализации.
						last_name VARCHAR(100), -- Фамилия пользователя.
						is_active BOOLEAN DEFAULT TRUE, -- Флаг активности пользователя. Можно деактивировать вместо полного удаления.
						role VARCHAR(50) DEFAULT 'user', -- Роль пользователя, например: 'user', 'admin'. Удобно для простых RBAC (role-bas
						created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Дата и время регистрации пользователя.
						updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);`
	InsertQuery = `INSERT INTO users (username, password_hash) VALUES ($1, $2)`
	SelectQuery = `SELECT id, username, password_hash FROM users WHERE username = $1`

	InitTasksQuery = `CREATE TABLE IF NOT EXISTS tasks (
							id SERIAL PRIMARY KEY,             -- Уникальный идентификатор задачи
							title TEXT NOT NULL,               -- Заголовок задачи
							description TEXT,                  -- Описание задачи (необязательное поле)
							status TEXT CHECK (status IN ('new', 'in_progress', 'done')) DEFAULT 'new', -- Статус задачи
							created_at TIMESTAMP DEFAULT now(), -- Время создания задачи
							updated_at TIMESTAMP DEFAULT now()  -- Время последнего обновления задачи
						);`
	insertTaskQuery = `INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id;`
	getTaskQuery    = `SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1;`
)
