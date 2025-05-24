-- Создание функции для удаления старых сообщений
CREATE OR REPLACE FUNCTION delete_old_messages()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM chat_messages
    WHERE timestamp < NOW() - INTERVAL '10 minutes';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создание триггера, который вызывает функцию после вставки нового сообщения
CREATE TRIGGER cleanup_old_messages
AFTER INSERT ON chat_messages
FOR EACH ROW
EXECUTE FUNCTION delete_old_messages();