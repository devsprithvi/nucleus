$version: "2"

namespace org.devsprithvi.nucleus.todo.domain

use org.devsprithvi.nucleus.todo#Task
use org.devsprithvi.nucleus.todo#TaskStatus

// ========================================================
// THE PORT (DATABASE INTERFACE)
// ========================================================
// We use "service" here because we want a clean interface 
// with a group of operations.
service TodoRepository {
    version: "1.0.0"
    operations: [
        Save,
        FindById,
        ListByStatus
    ]
}

// ========================================================
// PORT OPERATIONS (The Interface Methods)
// ========================================================

// 1. SAVE: The implementation will handle "Create" vs "Update" logic
operation Save {
    input: Task
    output: Task
}

// 2. FIND BY ID: Simple lookup
operation FindById {
    input: FindByIdInput
    output: Task
}

structure FindByIdInput {
    @required
    id: String
}

// 3. LIST: Filtering
operation ListByStatus {
    input: ListByStatusInput
    output: TaskList
}

structure ListByStatusInput {
    status: TaskStatus
}

// Helper List
list TaskList {
    member: Task
}