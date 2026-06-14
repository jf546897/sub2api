-- Mark unused Ops retry/replay storage as retired without destroying data.
-- The retry endpoints are no longer exposed, but already-captured request
-- context and retry audit rows may be needed for incident review. Keep this
-- migration intentionally non-destructive.

DO $$
BEGIN
  IF to_regclass('public.ops_error_logs') IS NOT NULL THEN
    COMMENT ON TABLE ops_error_logs IS 'Ops error logs (vNext). Stores sanitized error details; request replay fields are retained only for historical compatibility.';
  END IF;

  IF to_regclass('public.ops_retry_attempts') IS NOT NULL THEN
    COMMENT ON TABLE ops_retry_attempts IS 'Historical audit table for retired ops retry/replay flows; retained for compatibility and incident review.';
  END IF;
END $$;
