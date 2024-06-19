package types

import "AIComputingNode/pkg/protocol"

func AIProject2ProtocolMessage(projs []AIProjectOfNode) *protocol.AIProjectResponse {
	res := &protocol.AIProjectResponse{}
	for _, proj := range projs {
		res.Projects = append(res.Projects, &protocol.AIProjectOfNode{
			Project: proj.Project,
			Models:  proj.Models,
		})
	}
	return res
}

func ProtocolMessage2AIProject(res *protocol.AIProjectResponse) []AIProjectOfNode {
	projects := make([]AIProjectOfNode, len(res.Projects))
	for i, project := range res.Projects {
		projects[i].Project = project.Project
		projects[i].Models = project.Models
	}
	return projects
}
