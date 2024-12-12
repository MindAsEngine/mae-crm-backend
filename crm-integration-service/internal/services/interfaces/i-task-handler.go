package taskhandler

type TaskHandler interface {
    ProcessTask(taskID int) error
}
