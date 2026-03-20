CREATE TYPE payment_status AS ENUM ('pending', 'success', 'fail', 'cancel', 'timeout')

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    booking_id UUID,
    order_id UUID,
    status payment_status,
    price BIGINT NOT NULL,
    currency VARCHAR(50) NOT NULL
    transaction_id UUID
    payment_method VARCHAR(50)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);