# UI states and diagnostics

Every asynchronous or content-bearing view defines the following states before it
is considered complete:

- **Empty**: explains what the user can do next without showing inactive controls
  that have no implementation.
- **Loading**: identifies the current operation and keeps existing stable content
  visible when safe.
- **Partial**: shows available content and explicitly identifies failed or
  unavailable entries.
- **Success**: presents content, source provenance, and relevant engine status.
- **Failure**: uses plain language, preserves the workspace, and offers a valid
  retry or alternative where possible.
- **Cancelled**: is distinct from failure and does not display a misleading error.
- **Offline/unavailable**: applies only when an actual external dependency exists.

Diagnostics have severity, source, artifact or class context, and an actionable
summary. Engine console output is available in details but is not the primary
error presentation. Logs must not silently include private source or credentials.

Long-running operations expose progress when measurable and always support
cancellation. Closing a tab or archive cancels its owned work. Repeated actions
must not create duplicate tasks or stale result updates.

Tests cover every state that the implementation can produce. Screenshots for
visual review include empty, loading, failure, and populated states where those
states change layout.
