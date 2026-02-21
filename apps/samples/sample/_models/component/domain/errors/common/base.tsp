namespace Org.DevsPrithvi.Nucleus.Error.Common;

/**
 * The Root Error (The Source of Truth)
 * 
 * This model serves as the base for all error responses in the system.
 * It ensures a consistent structure for error reporting across all domains.
 */
@error
@doc("The base error model for all system errors.")
model BaseError {
  /**
   * A human-readable message describing the error.
   * This message is intended for developers or to be shown in logs.
   */
  @doc("Human-readable description of the error.")
  message: string;

  /**
   * A machine-readable error code.
   * Useful for client-side logic to identify specific error types without parsing the message.
   */
  @doc("Machine-readable error code.")
  code: string;
}

/**
 * The Semantic Separation: Client Faults
 * 
 * We create these purely to categorise the faults logically.
 * Logic: "This is the caller's fault" (e.g., bad input, unauthorized access).
 * NO HTTP CODES here. Just logical grouping.
 */
@error
@doc("Base model for errors caused by the client's request.")
model ClientError extends BaseError {
  // Concern: Client-side issues that can be fixed by the caller.
}

/**
 * The Semantic Separation: Server Faults
 * 
 * We create these purely to categorise the faults logically.
 * Logic: "This is the system's fault" (e.g., database down, unhandled exception).
 * NO HTTP CODES here. Just logical grouping.
 */
@error
@doc("Base model for errors caused by internal system issues.")
model ServerError extends BaseError {
  /**
   * Suggests how long the client should wait before retrying the request.
   */
  @doc("The number of seconds after which the client may retry.")
  retryAfterSeconds?: int32;
}

// --- Concrete Client Errors ---

/**
 * Thrown when the provided arguments are invalid or do not meet validation criteria.
 */
@doc("The request contains invalid arguments.")
model InvalidArgument extends ClientError {
  /**
   * The name of the field that failed validation.
   */
  @doc("The field that failed validation.")
  invalidField: string;
}

/**
 * Thrown when a requested resource (e.g., a User or Order) does not exist.
 */
@doc("The requested resource was not found.")
model ResourceNotFound extends ClientError {
  /**
   * The unique identifier of the resource that was not found.
   */
  @doc("The ID of the missing resource.")
  resourceId: string;
}

/**
 * Thrown when the caller is authenticated but lacks permission to access the resource.
 */
@doc("Permission denied for the requested operation.")
model AccessDenied extends ClientError {}

/**
 * Thrown when the caller is not authenticated.
 */
@doc("The request lacks valid authentication credentials.")
model Unauthenticated extends ClientError {}

/**
 * Thrown when the operation conflicts with the current state of the resource.
 */
@doc("The operation conflicted with the current state of the resource.")
model ResourceConflict extends ClientError {
  /**
   * The unique identifier of the resource that is in conflict.
   */
  @doc("The ID of the resource in conflict.")
  resourceId: string;
}

// --- Concrete Server Errors ---

/**
 * A generic error indicating something went wrong on the server.
 */
@doc("An unexpected internal server error occurred.")
model InternalError extends ServerError {}

/**
 * Thrown when the service is temporarily unable to handle the request.
 */
@doc("The service is temporarily unavailable.")
model ServiceUnavailable extends ServerError {}