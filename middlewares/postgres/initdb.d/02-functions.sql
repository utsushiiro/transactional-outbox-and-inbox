CREATE OR REPLACE FUNCTION auto_set_created_at_and_updated_at() 
  RETURNS trigger AS $$
  BEGIN
    IF TG_OP = 'INSERT' THEN
      NEW.created_at := NOW();
      NEW.updated_at := NOW();
    ELSIF TG_OP = 'UPDATE' THEN
      NEW.created_at := OLD.created_at;
      NEW.updated_at := NOW();
    END IF;

    RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;
