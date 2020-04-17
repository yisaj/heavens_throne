UPDATE player SET location=NULL, next_location=NULL WHERE dead=TRUE;
ALTER TABLE player DROP COLUMN dead;