package taskhandler

type TaskHandler interface {
    ProcessUploadMsg(UploadMsgID int) error
}
