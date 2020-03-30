CREATE OR REPLACE FUNCTION update_content_analytics(id varchar, views int) RETURNS VOID AS $$ 
  DECLARE 
  BEGIN 
      UPDATE content_analytics 
        SET pageviews = pageviews + views
        WHERE content=id; 
      IF NOT FOUND THEN 
      INSERT INTO content_analytics values (id, views); 
      END IF; 
  END; 
  $$ LANGUAGE 'plpgsql'; 