package domain

// ProjectStatus is the lifecycle state of a long-lived project context.
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

// ProjectContextRelation identifies how one project context node relates to another.
type ProjectContextRelation string

const (
	ProjectContextRelationSupports  ProjectContextRelation = "supports"
	ProjectContextRelationBlocks    ProjectContextRelation = "blocks"
	ProjectContextRelationRefines   ProjectContextRelation = "refines"
	ProjectContextRelationDependsOn ProjectContextRelation = "depends_on"
	ProjectContextRelationRelated   ProjectContextRelation = "related"
)

// Actor identifies who created or changed project context (mirrors tasks/domain.Actor wire values).
type Actor string

const (
	ActorUser  Actor = "user"
	ActorAgent Actor = "agent"
)
