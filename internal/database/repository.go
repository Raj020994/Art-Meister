package database

import (
	"context"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error)
	GetUser(ctx context.Context, id uuid.UUID) (GetUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
	PatchUserImages(ctx context.Context, arg PatchUserImagesParams) (PatchUserImagesRow, error)
	PatchUserProfile(ctx context.Context, arg PatchUserProfileParams) (PatchUserProfileRow, error)
	PatchUserAdmin(ctx context.Context, arg PatchUserAdminParams) (PatchUserAdminRow, error)
	PatchUserPassword(ctx context.Context, arg PatchUserPasswordParams) (PatchUserPasswordRow, error)
}
type ArtRepository interface {
	DeleteArt(ctx context.Context, arg DeleteArtParams) (uuid.UUID, error)
	GetArtByID(ctx context.Context, id uuid.UUID) (Art, error)
	GetArtByUser(ctx context.Context, userID uuid.UUID) ([]Art, error)
	ListArt(ctx context.Context) ([]Art, error)
	ListArtByTag(ctx context.Context, tags []string) ([]Art, error)
	ListArtByTags(ctx context.Context, dollar_1 []string) ([]Art, error)
	UpdateArt(ctx context.Context, arg UpdateArtParams) (Art, error)
	UpdateArtStatus(ctx context.Context, arg UpdateArtStatusParams) (Art, error)
	CreateArt(ctx context.Context, arg CreateArtParams) (Art, error)
}
type ArtMetaDataRepository interface {
	AddArtComment(ctx context.Context, arg AddArtCommentParams) (ArtComment, error)
	CheckArtLikedByUser(ctx context.Context, arg CheckArtLikedByUserParams) (bool, error)
	DeleteArtComment(ctx context.Context, arg DeleteArtCommentParams) (uuid.UUID, error)
	GetArtCommentsByArtID(ctx context.Context, artID uuid.UUID) ([]GetArtCommentsByArtIDRow, error)
	GetArtCommentsCount(ctx context.Context, artID uuid.UUID) (int32, error)
	GetArtLikesCount(ctx context.Context, artID uuid.UUID) (int32, error)
	LikeArt(ctx context.Context, arg LikeArtParams) (ArtLike, error)
	UnlikeArt(ctx context.Context, arg UnlikeArtParams) (uuid.UUID, error)
}
type EventRepository interface {
	CreateEvent(ctx context.Context, arg CreateEventParams) (Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (Event, error)
	ListEvents(ctx context.Context) ([]Event, error)
	ListEventsByMode(ctx context.Context, status ModeOfConduct) ([]Event, error)
	ListUpcomingEvents(ctx context.Context) ([]Event, error)
	UpdateEvent(ctx context.Context, arg UpdateEventParams) (Event, error)
}
type EventAttendeesRepository interface {
	CountEventAttendees(ctx context.Context, eventID uuid.UUID) (int32, error)
	EnrollUserToEvent(ctx context.Context, arg EnrollUserToEventParams) (EventAttendee, error)
	ListEventAttendees(ctx context.Context, eventID uuid.UUID) ([]User, error)
	ListMyEvents(ctx context.Context, userID uuid.UUID) ([]Event, error)
	RemoveUserFromEvent(ctx context.Context, arg RemoveUserFromEventParams) (uuid.UUID, error)
}
