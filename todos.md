# TODO: For authentication

- Move Dex redirect to backend instead of frontend, to avoid exposing the Dex client secret in the frontend code

# TODO: This should be moved to issues or discarded based

- Pagingation handling for RabbitMQ and Prometheus API calls, to ensure all data is retrieved when there are more than 100 results
- Seeding should use the same logic as the API, to help ensure that the data is consistent and valid
- Automatic cleanup of old alarms, to prevent the database from growing indefinitely
- Automatic cleanup of old notification vhosts, to prevent the database from growing indefinitely
- Better messaging to webhook and email receivers when an alarm is triggered, provide more context and information about the alarm
- Make email and webhook receivers unique per rule, duplicate receivers shouldn't be allowed

## Low priority ToDo's:

- Retry logic for connecting to RabbitMQ and Prometheus, to handle transient connectivity issues
- Watchers for mongodb, rabbitmq, prometheus to cover connectivity loss
- More comprehensive unit tests for all packages, especially the notify package
- Checker does unnecessary updates, should only update when there are changes to the alarm state (e.g., from not triggered to triggered, or vice versa)
