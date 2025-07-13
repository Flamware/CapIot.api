ALTER TABLE public.sensors
    ADD COLUMN min_threshold NUMERIC(10, 2),
    ADD COLUMN max_threshold NUMERIC(10, 2);