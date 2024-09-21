CREATE TABLE IF NOT EXISTS public.user (
    id UUID UNIQUE PRIMARY KEY DEFAULT (gen_random_uuid()),
    email TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_user_id ON public.user(id);

CREATE TABLE IF NOT EXISTS refresh (
    refresh_id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.user(id)
);

CREATE INDEX IF NOT EXISTS idx_refresh_user_id ON refresh(user_id);