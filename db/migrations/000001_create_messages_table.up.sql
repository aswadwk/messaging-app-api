CREATE TABLE messages (
  id UUID,
  tenant_id UUID NOT NULL,
  payload JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (id, tenant_id)
) PARTITION BY LIST (tenant_id);