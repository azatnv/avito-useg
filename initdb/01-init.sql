CREATE TABLE users (
    id int PRIMARY KEY
);

CREATE TABLE segments (
    id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE user2seg (
    u_id int,
    s_id int,
    date_add TIMESTAMP NOT NULL,
    date_end TIMESTAMP NOT NULL,
    CONSTRAINT pk_user2seg PRIMARY KEY (u_id, s_id),
    CONSTRAINT fk_user2seg_1 FOREIGN KEY (u_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_user2seg_2 FOREIGN KEY (s_id) REFERENCES segments(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE actions (
    id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    u_id int NOT NULL,
    s_id int NOT NULL,
    action TEXT NOT NULL,
    date TIMESTAMP NOT NULL,
    CONSTRAINT fk_actions_1 FOREIGN KEY (u_id) REFERENCES users(id) ON DELETE NO ACTION ON UPDATE CASCADE,
    CONSTRAINT fk_actions_2 FOREIGN KEY (s_id) REFERENCES segments(id) ON DELETE NO ACTION ON UPDATE CASCADE
);

CREATE FUNCTION user2seg_register_insert_action() RETURNS trigger AS
$$BEGIN
    INSERT INTO actions (u_id, s_id, action, date)
        VALUES (NEW.u_id, NEW.s_id, 'add', NEW.date_add);
    RETURN NEW;
END;$$ LANGUAGE plpgsql;

CREATE TRIGGER user2seg_register_insert_action AFTER INSERT ON user2seg
    FOR EACH ROW EXECUTE PROCEDURE user2seg_register_insert_action();

CREATE FUNCTION user2seg_register_delete_action() RETURNS trigger AS
$$BEGIN
    INSERT INTO actions (u_id, s_id, action, date)
        VALUES (OLD.u_id, OLD.s_id, 'remove', current_timestamp(0));
    RETURN OLD;
END;$$ LANGUAGE plpgsql;

CREATE TRIGGER user2seg_register_delete_action AFTER DELETE ON user2seg
    FOR EACH ROW EXECUTE PROCEDURE user2seg_register_delete_action();
