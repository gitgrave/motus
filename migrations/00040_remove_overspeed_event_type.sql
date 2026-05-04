-- +goose Up

-- Remove the overspeed event type. The OverspeedService is being deleted
-- because we don't have a reliable data source for per-device speed limits.
-- Hardware overspeed alarms reported by trackers (e.g. H02 alarm bit 2) are
-- still emitted as 'alarm' events with attributes.alarm = 'overspeed'.
DELETE FROM events WHERE type = 'overspeed';
ALTER TABLE events DROP CONSTRAINT IF EXISTS valid_event_type;
ALTER TABLE events ADD CONSTRAINT valid_event_type
CHECK (type IN ('geofenceEnter', 'geofenceExit', 'deviceOnline', 'deviceOffline', 'motion', 'deviceIdle', 'ignitionOn', 'ignitionOff', 'alarm', 'tripCompleted'));

-- +goose Down
ALTER TABLE events DROP CONSTRAINT IF EXISTS valid_event_type;
ALTER TABLE events ADD CONSTRAINT valid_event_type
CHECK (type IN ('geofenceEnter', 'geofenceExit', 'deviceOnline', 'deviceOffline', 'overspeed', 'motion', 'deviceIdle', 'ignitionOn', 'ignitionOff', 'alarm', 'tripCompleted'));
