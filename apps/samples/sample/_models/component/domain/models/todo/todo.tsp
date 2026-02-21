namespace Org.DevsPrithvi.Nucleus.Todo;

// 1. Reusable Components
model Identifiable {
  id: string;
}

model Auditable {
  createdAt: utcDateTime;
  updatedAt: utcDateTime;
}

// 2. The Domain Value (Enum)
enum TaskStatus {
  Todo,
  InProgress,
  Done
}

// 3. The Aggregate Root (The "Task")
// STRATEGY: Use "..." to achieve Multiple Composition
model Task {
  // Mixin 1: Identity
  ...Identifiable;

  // Mixin 2: Audit trails
  ...Auditable;

  // Task Specifics
  title: string;
  description?: string;
  status: TaskStatus;
}