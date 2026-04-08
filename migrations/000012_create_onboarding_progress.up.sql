CREATE TABLE onboarding_progress (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    cli_installed BOOLEAN NOT NULL DEFAULT false,
    first_push    BOOLEAN NOT NULL DEFAULT false,
    second_device BOOLEAN NOT NULL DEFAULT false,
    completed     BOOLEAN NOT NULL DEFAULT false,
    dismissed     BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
