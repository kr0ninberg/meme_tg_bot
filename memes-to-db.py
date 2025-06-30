import os
import psycopg2

# Путь к папке с картинками
MEMES_DIR = './memes'

# Подключение к базе данных
conn = psycopg2.connect( # WARNING credentials in code!
    dbname='memes',
    user='memes_user',
    password='memes_pass',
    host='localhost',
    port='5432'
)

cursor = conn.cursor()

# Получаем список файлов в папке
for filename in os.listdir(MEMES_DIR):
    if os.path.isfile(os.path.join(MEMES_DIR, filename)):
        name, ext = os.path.splitext(filename)
        ext = ext.lstrip('.').lower()  # Удаляем точку и приводим к нижнему регистру
        if ext in ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp']:
            cursor.execute(
                """
                INSERT INTO memes_filenames (filename, format)
                VALUES (%s, %s)
                ON CONFLICT (filename) DO NOTHING;
                """,
                (filename, ext)
            )

# Сохраняем изменения и закрываем соединение
conn.commit()
cursor.close()
conn.close()