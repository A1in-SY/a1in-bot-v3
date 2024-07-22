package api

import "a1in-bot-v3/model/segment"

func BuildSendPrivateMsgRequest(echo string, userId int64, segs ...*segment.Segment) *APIRequest {
	return &APIRequest{
		Action: "send_private_msg",
		Params: &APIRequestParams{
			UserId:  userId,
			Message: segs,
		},
		Echo: echo,
	}
}

func BuildSendGroupMsgRequest(echo string, groupId int64, segs ...*segment.Segment) *APIRequest {
	return &APIRequest{
		Action: "send_group_msg",
		Params: &APIRequestParams{
			GroupId: groupId,
			Message: segs,
		},
		Echo: echo,
	}
}

func BuildUploadGroupFileRequest(echo string, groupId int64, file, name string) *APIRequest {
	return &APIRequest{
		Action: "upload_group_file",
		Params: &APIRequestParams{
			GroupId: groupId,
			File:    file,
			Name:    name,
			Folder:  "",
		},
		Echo: echo,
	}
}
