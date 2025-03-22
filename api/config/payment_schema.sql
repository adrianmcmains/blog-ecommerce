-- Payment Transactions Table
CREATE TABLE IF NOT EXISTS payment_transactions (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payment_url TEXT,
    reference VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on transaction_id
CREATE INDEX IF NOT EXISTS idx_payment_transactions_transaction_id ON payment_transactions(transaction_id);

-- Add transaction_id to orders table if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'orders' AND column_name = 'transaction_id'
    ) THEN
        ALTER TABLE orders ADD COLUMN transaction_id VARCHAR(255);
    END IF;
END
$$;

-- Ensure orders table has appropriate status values for payments
DO $$
BEGIN
    -- Create a temporary function to check if ENUM type exists and has required values
    CREATE OR REPLACE FUNCTION temp_check_order_status() RETURNS VOID AS $$
    DECLARE
        status_check TEXT;
    BEGIN
        -- Check if the column is of type varchar/text or enum
        SELECT data_type INTO status_check 
        FROM information_schema.columns 
        WHERE table_name = 'orders' AND column_name = 'status';

        IF status_check = 'character varying' OR status_check = 'text' THEN
            -- Column is VARCHAR/TEXT, no need for enum operations
            RETURN;
        ELSIF status_check LIKE '%enum%' THEN
            -- Column is an ENUM, check/add values
            -- This is a simplified approach, as proper enum modification would be more complex
            -- For real applications, consider using a migration tool or more robust approach
            BEGIN
                -- Try to add new values - this will fail if they already exist, which is fine
                ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'payment_pending';
                ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'paid';
                ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'payment_failed';
                ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'payment_canceled';
            EXCEPTION WHEN OTHERS THEN
                -- Ignore errors, as they likely mean values already exist
                NULL;
            END;
        END IF;
    END;
    $$ LANGUAGE plpgsql;

    -- Execute the function
    PERFORM temp_check_order_status();

    -- Drop the temporary function
    DROP FUNCTION temp_check_order_status();
END
$$;

-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_payment_transactions_order_id ON payment_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_created_at ON payment_transactions(created_at);

-- Create a view for payment reporting
CREATE OR REPLACE VIEW payment_reports AS
SELECT 
    p.id,
    p.transaction_id,
    p.provider,
    p.amount,
    p.currency,
    p.status,
    p.reference,
    o.id as order_id,
    u.id as user_id,
    u.email as user_email,
    u.name as user_name,
    p.created_at,
    p.updated_at
FROM payment_transactions p
JOIN orders o ON p.order_id = o.id
JOIN users u ON o.user_id = u.id;