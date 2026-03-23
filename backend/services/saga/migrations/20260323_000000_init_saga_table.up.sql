CREATE TYPE saga_status AS ENUM('PENDING', 'WAIT_FOR_PAYMENT' 'PROCESSING', 'COMPLETED', 'ROLLING_BACK', 'ROLLED_BACK', 'FAIL')



CREATE TABLE IF NOT EXISTS sagas(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID UUID NOT NULL,
    name NVARCHAR(100) NOT NULL
    status saga_status,
    current_step_index INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
)

CREATE TYPE step_status AS ENUM ('PENDING', 'EXECUTING', 'COMPLETED','COMPENSATING', 'COMPENSATED','FAILED')

CREATE TABLE IF NOT EXISTS saga_step(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    saga_id UUID NOT NULL REFERENCES saga(id) ON DELETE CASCADE
    name NVARCHAR(100) NOT NULL
    executed_at TIMESTAMPTZ
    compensated_at TIMESTAMPTZ
    status step_status
    order INTEGER
    should_pause_for_payment BOOLEAN
)