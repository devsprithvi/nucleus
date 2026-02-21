import "../common/base.tsp";

namespace Org.DevsPrithvi.Nucleus.Error.Todo;

using Org.DevsPrithvi.Nucleus.Error.Common;

// 1. Task Limit Exceeded
@doc("Thrown when a user has reached the maximum number of tasks allowed.")
model TaskLimitExceeded extends ClientError {
  @doc("The maximum number of tasks allowed.")
  limit: int32;
}

// 2. Invalid Transition
@doc("Thrown when an invalid status transition is attempted on a task.")
model InvalidTaskStatusTransition extends ClientError {
  @doc("The current status of the task.")
  currentStatus: string;

  @doc("The requested status that is invalid for the current state.")
  requestedStatus: string;
}

// 3. Task Already Completed
@doc("Thrown when an operation is attempted on a task that is already completed.")
model TaskAlreadyCompleted extends ClientError {
  @doc("The ID of the task that is already completed.")
  taskId: string;
}