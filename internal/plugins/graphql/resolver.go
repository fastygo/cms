package graphqlplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	graphql "github.com/graph-gophers/graphql-go"

	appcontent "github.com/fastygo/cms/internal/application/content"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmeta "github.com/fastygo/cms/internal/application/meta"
	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/framework/pkg/web/locale"
)

type rootResolver struct {
	services     Services
	registry     *plugins.Registry
	metaRegistry *appmeta.Registry
}

type contentListArgs struct {
	Page     int32
	PerPage  int32
	Status   *[]string
	Author   *string
	Taxonomy *string
	Term     *string
	Search   *string
	Locale   *string
	After    *graphql.Time
	Before   *graphql.Time
	Sort     *string
	Order    *string
}

type contentLookupArgs struct {
	ID     *graphql.ID
	Slug   *string
	Locale *string
}

type searchArgs struct {
	Query   string
	Page    int32
	PerPage int32
	Locale  *string
}

type termsArgs struct {
	Type string
}

type mediaArgs struct {
	ID *graphql.ID
}

type menusArgs struct {
	Location *string
}

type contentMutationArgs struct {
	Input contentInput
}

type updateContentArgs struct {
	ID    graphql.ID
	Input contentInput
}

type contentIDArgs struct {
	ID graphql.ID
}

type scheduleArgs struct {
	ID          graphql.ID
	PublishedAt graphql.Time
}

type assignTermsArgs struct {
	ContentID graphql.ID
	Terms     []termRefInput
}

type attachMediaArgs struct {
	ContentID graphql.ID
	AssetID   graphql.ID
}

type saveMenuArgs struct {
	Input menuInput
}

type saveSettingArgs struct {
	Input settingInput
}

type contentInput struct {
	Status          *string
	Title           JSONValue
	Slug            JSONValue
	Content         JSONValue
	Excerpt         *JSONValue
	AuthorID        *string
	FeaturedMediaID *string
	Template        *string
	Metadata        *JSONValue
	Terms           *[]termRefInput
	PublishedAt     *graphql.Time
}

type termRefInput struct {
	Taxonomy string
	TermID   string
}

type menuInput struct {
	ID       string
	Name     string
	Location string
	Items    []menuItemInput
}

type menuItemInput struct {
	ID       string
	Label    string
	URL      string
	Kind     *string
	TargetID *string
	Children *[]menuItemInput
}

type settingInput struct {
	Key    string
	Value  JSONValue
	Public bool
}

type JSONValue struct {
	Value any
}

func (JSONValue) ImplementsGraphQLType(name string) bool {
	return name == "JSON"
}

func (j JSONValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Value)
}

func (j *JSONValue) UnmarshalGraphQL(input any) error {
	j.Value = input
	return nil
}

func (r *rootResolver) Posts(ctx context.Context, args contentListArgs) (*contentListResolver, error) {
	query, err := buildContentQuery(args, domaincontent.KindPost, publicOnly(ctx))
	if err != nil {
		return nil, err
	}
	result, err := r.services.Content.List(ctx, query)
	if err != nil {
		return nil, err
	}
	return r.newContentListResolver(ctx, result), nil
}

func (r *rootResolver) Post(ctx context.Context, args contentLookupArgs) (*contentResolver, error) {
	return r.lookupContent(ctx, domaincontent.KindPost, args)
}

func (r *rootResolver) Pages(ctx context.Context, args contentListArgs) (*contentListResolver, error) {
	query, err := buildContentQuery(args, domaincontent.KindPage, publicOnly(ctx))
	if err != nil {
		return nil, err
	}
	result, err := r.services.Content.List(ctx, query)
	if err != nil {
		return nil, err
	}
	return r.newContentListResolver(ctx, result), nil
}

func (r *rootResolver) Page(ctx context.Context, args contentLookupArgs) (*contentResolver, error) {
	return r.lookupContent(ctx, domaincontent.KindPage, args)
}

func (r *rootResolver) ContentTypes(ctx context.Context) ([]*contentTypeResolver, error) {
	items, err := r.services.ContentTypes.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*contentTypeResolver, 0, len(items))
	for _, item := range items {
		if !item.GraphQLVisible {
			continue
		}
		item := item
		out = append(out, &contentTypeResolver{item: item})
	}
	return out, nil
}

func (r *rootResolver) Taxonomies(ctx context.Context) ([]*taxonomyResolver, error) {
	items, err := r.services.Taxonomy.ListDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*taxonomyResolver, 0, len(items))
	for _, item := range items {
		if !item.GraphQLVisible {
			continue
		}
		item := item
		out = append(out, &taxonomyResolver{item: item})
	}
	return out, nil
}

func (r *rootResolver) Terms(ctx context.Context, args termsArgs) ([]*termResolver, error) {
	definition, ok, err := r.services.Taxonomy.GetDefinition(ctx, domaintaxonomy.Type(args.Type))
	if err != nil {
		return nil, err
	}
	if !ok || !definition.GraphQLVisible {
		return nil, nil
	}
	items, err := r.services.Taxonomy.ListTerms(ctx, domaintaxonomy.Type(args.Type))
	if err != nil {
		return nil, err
	}
	out := make([]*termResolver, 0, len(items))
	for _, item := range items {
		item := item
		out = append(out, &termResolver{item: item})
	}
	return out, nil
}

func (r *rootResolver) Media(ctx context.Context, args mediaArgs) ([]*mediaResolver, error) {
	if args.ID != nil {
		item, ok, err := r.services.Media.Get(ctx, domainmedia.ID(*args.ID))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
		return []*mediaResolver{{item: item}}, nil
	}
	items, err := r.services.Media.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*mediaResolver, 0, len(items))
	for _, item := range items {
		item := item
		out = append(out, &mediaResolver{item: item})
	}
	return out, nil
}

func (r *rootResolver) Authors(ctx context.Context) ([]*authorResolver, error) {
	users, err := r.services.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*authorResolver, 0, len(users))
	for _, user := range users {
		if user.Status != domainusers.StatusActive {
			continue
		}
		profile := user.PublicAuthor()
		out = append(out, &authorResolver{profile: profile, media: r.services.Media})
	}
	return out, nil
}

func (r *rootResolver) Menus(ctx context.Context, args menusArgs) ([]*menuResolver, error) {
	if args.Location != nil && strings.TrimSpace(*args.Location) != "" {
		menu, ok, err := r.services.Menus.ByLocation(ctx, domainmenus.Location(*args.Location))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
		return []*menuResolver{{item: menu}}, nil
	}
	items, err := r.services.Menus.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*menuResolver, 0, len(items))
	for _, item := range items {
		item := item
		out = append(out, &menuResolver{item: item})
	}
	return out, nil
}

func (r *rootResolver) Settings(ctx context.Context) ([]*settingResolver, error) {
	items, err := r.services.Settings.Public(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*settingResolver, 0, len(items))
	for _, item := range items {
		item := item
		out = append(out, &settingResolver{item: item})
	}
	return out, nil
}

func (r *rootResolver) Search(ctx context.Context, args searchArgs) (*contentListResolver, error) {
	page := int(args.Page)
	if page <= 0 {
		page = 1
	}
	perPage := int(args.PerPage)
	if perPage <= 0 {
		perPage = 20
	}
	query := domaincontent.Query{
		Kinds:      []domaincontent.Kind{domaincontent.KindPost, domaincontent.KindPage},
		Search:     strings.TrimSpace(args.Query),
		Locale:     stringValue(args.Locale),
		PublicOnly: publicOnly(ctx),
		Page:       page,
		PerPage:    perPage,
	}
	result, err := r.services.Content.List(ctx, query)
	if err != nil {
		return nil, err
	}
	return r.newContentListResolver(ctx, result), nil
}

func (r *rootResolver) CreatePost(ctx context.Context, args contentMutationArgs) (*contentResolver, error) {
	return r.createContent(ctx, domaincontent.KindPost, args.Input)
}

func (r *rootResolver) CreatePage(ctx context.Context, args contentMutationArgs) (*contentResolver, error) {
	return r.createContent(ctx, domaincontent.KindPage, args.Input)
}

func (r *rootResolver) UpdateContent(ctx context.Context, args updateContentArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Content.Update(ctx, principal, appcontent.UpdateCommand{
		ID:              domaincontent.ID(args.ID),
		Title:           localizedFromJSON(args.Input.Title, false),
		Slug:            localizedFromJSON(args.Input.Slug, false),
		Body:            localizedFromJSON(args.Input.Content, false),
		Excerpt:         optionalLocalized(args.Input.Excerpt),
		AuthorID:        stringValue(args.Input.AuthorID),
		FeaturedMediaID: stringValue(args.Input.FeaturedMediaID),
		Template:        stringValue(args.Input.Template),
		Metadata:        metadataFromJSON(args.Input.Metadata),
		Terms:           termRefsFromInput(args.Input.Terms),
	})
	if err != nil {
		return nil, err
	}
	entry, err = r.applyRequestedStatus(ctx, principal, entry.ID, args.Input)
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) PublishContent(ctx context.Context, args contentIDArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Content.Publish(ctx, principal, domaincontent.ID(args.ID))
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) ScheduleContent(ctx context.Context, args scheduleArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Content.Schedule(ctx, principal, domaincontent.ID(args.ID), args.PublishedAt.Time)
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) TrashContent(ctx context.Context, args contentIDArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Content.Trash(ctx, principal, domaincontent.ID(args.ID))
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) RestoreContent(ctx context.Context, args contentIDArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Content.Restore(ctx, principal, domaincontent.ID(args.ID))
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) AssignTerms(ctx context.Context, args assignTermsArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Taxonomy.AssignTerms(ctx, principal, domaincontent.ID(args.ContentID), termRefsFromSlice(args.Terms))
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) AttachFeaturedMedia(ctx context.Context, args attachMediaArgs) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Media.AttachFeatured(ctx, principal, domaincontent.ID(args.ContentID), domainmedia.ID(args.AssetID))
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) SaveMenu(ctx context.Context, args saveMenuArgs) (*menuResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	menu := domainmenus.Menu{
		ID:       domainmenus.ID(args.Input.ID),
		Name:     args.Input.Name,
		Location: domainmenus.Location(args.Input.Location),
		Items:    menuItemsFromInput(args.Input.Items),
	}
	if err := r.services.Menus.Save(ctx, principal, menu); err != nil {
		return nil, err
	}
	return &menuResolver{item: menu}, nil
}

func (r *rootResolver) SaveSetting(ctx context.Context, args saveSettingArgs) (*settingResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	value := domainsettings.Value{
		Key:    domainsettings.Key(args.Input.Key),
		Value:  args.Input.Value.Value,
		Public: args.Input.Public,
	}
	if err := r.services.Settings.Save(ctx, principal, value); err != nil {
		return nil, err
	}
	return &settingResolver{item: value}, nil
}

func (r *rootResolver) createContent(ctx context.Context, kind domaincontent.Kind, input contentInput) (*contentResolver, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := r.services.Content.CreateDraft(ctx, principal, appcontent.CreateDraftCommand{
		Kind:            kind,
		Title:           localizedFromJSON(input.Title, false),
		Slug:            localizedFromJSON(input.Slug, false),
		Body:            localizedFromJSON(input.Content, false),
		Excerpt:         optionalLocalized(input.Excerpt),
		AuthorID:        stringValue(input.AuthorID),
		FeaturedMediaID: stringValue(input.FeaturedMediaID),
		Template:        stringValue(input.Template),
		Metadata:        metadataFromJSON(input.Metadata),
		Terms:           termRefsFromInput(input.Terms),
	})
	if err != nil {
		return nil, err
	}
	entry, err = r.applyRequestedStatus(ctx, principal, entry.ID, input)
	if err != nil {
		return nil, err
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) lookupContent(ctx context.Context, kind domaincontent.Kind, args contentLookupArgs) (*contentResolver, error) {
	principal := principalFromContext(ctx)
	locale := stringValue(args.Locale)
	var (
		entry domaincontent.Entry
		err   error
	)
	switch {
	case args.ID != nil:
		entry, err = r.services.Content.Get(ctx, principal, domaincontent.ID(*args.ID))
	case args.Slug != nil && strings.TrimSpace(*args.Slug) != "":
		entry, err = r.services.Content.GetBySlug(ctx, principal, kind, *args.Slug, locale)
	default:
		return nil, fmt.Errorf("either id or slug is required")
	}
	if err != nil {
		if hidesExistence(err) {
			return nil, nil
		}
		return nil, err
	}
	if entry.Kind != kind {
		return nil, nil
	}
	return r.newContentResolver(ctx, entry), nil
}

func (r *rootResolver) applyRequestedStatus(ctx context.Context, principal domainauthz.Principal, id domaincontent.ID, input contentInput) (domaincontent.Entry, error) {
	status := strings.TrimSpace(stringValue(input.Status))
	switch domaincontent.Status(status) {
	case "", domaincontent.StatusDraft:
		return r.services.Content.Get(ctx, principal, id)
	case domaincontent.StatusPublished:
		return r.services.Content.Publish(ctx, principal, id)
	case domaincontent.StatusScheduled:
		if input.PublishedAt == nil {
			return domaincontent.Entry{}, fmt.Errorf("publishedAt is required for scheduled content")
		}
		return r.services.Content.Schedule(ctx, principal, id, input.PublishedAt.Time)
	case domaincontent.StatusTrashed:
		return r.services.Content.Trash(ctx, principal, id)
	default:
		return domaincontent.Entry{}, fmt.Errorf("unsupported content status %q", status)
	}
}

func (r *rootResolver) newContentListResolver(ctx context.Context, result domaincontent.ListResult) *contentListResolver {
	items := make([]*contentResolver, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, r.newContentResolver(ctx, item))
	}
	return &contentListResolver{
		items: items,
		pagination: paginationResolver{
			page:       result.Page,
			perPage:    result.PerPage,
			total:      result.Total,
			totalPages: result.TotalPages,
		},
	}
}

func (r *rootResolver) newContentResolver(ctx context.Context, entry domaincontent.Entry) *contentResolver {
	return &contentResolver{
		item:           entry,
		includePrivate: !publicOnly(ctx),
		registry:       r.registry,
		metaRegistry:   r.metaRegistry,
	}
}

type contentListResolver struct {
	items      []*contentResolver
	pagination paginationResolver
}

func (r *contentListResolver) Items() []*contentResolver {
	return r.items
}

func (r *contentListResolver) Pagination() *paginationResolver {
	return &r.pagination
}

type paginationResolver struct {
	page       int
	perPage    int
	total      int
	totalPages int
}

func (r *paginationResolver) Page() int32 {
	return int32(r.page)
}

func (r *paginationResolver) PerPage() int32 {
	return int32(r.perPage)
}

func (r *paginationResolver) Total() int32 {
	return int32(r.total)
}

func (r *paginationResolver) TotalPages() int32 {
	return int32(r.totalPages)
}

type contentResolver struct {
	item           domaincontent.Entry
	includePrivate bool
	registry       *plugins.Registry
	metaRegistry   *appmeta.Registry
}

func (r *contentResolver) ID() graphql.ID {
	return graphql.ID(r.item.ID)
}

func (r *contentResolver) Kind() string {
	return string(r.item.Kind)
}

func (r *contentResolver) Status() string {
	return string(r.item.Status)
}

func (r *contentResolver) Visibility() string {
	return string(r.item.Visibility)
}

func (r *contentResolver) Slug(ctx context.Context) (JSONValue, error) {
	projection, err := r.projection(ctx)
	if err != nil {
		return JSONValue{}, err
	}
	return JSONValue{Value: projection.Slug}, nil
}

func (r *contentResolver) Title(ctx context.Context) (JSONValue, error) {
	projection, err := r.projection(ctx)
	if err != nil {
		return JSONValue{}, err
	}
	return JSONValue{Value: projection.Title}, nil
}

func (r *contentResolver) Content(ctx context.Context) (JSONValue, error) {
	projection, err := r.projection(ctx)
	if err != nil {
		return JSONValue{}, err
	}
	return JSONValue{Value: projection.Content}, nil
}

func (r *contentResolver) Excerpt(ctx context.Context) (JSONValue, error) {
	projection, err := r.projection(ctx)
	if err != nil {
		return JSONValue{}, err
	}
	return JSONValue{Value: projection.Excerpt}, nil
}

func (r *contentResolver) AuthorID() string {
	return r.item.AuthorID
}

func (r *contentResolver) FeaturedMediaID() *string {
	if strings.TrimSpace(r.item.FeaturedMediaID) == "" {
		return nil
	}
	value := r.item.FeaturedMediaID
	return &value
}

func (r *contentResolver) Taxonomies() []*termAssignmentResolver {
	out := make([]*termAssignmentResolver, 0, len(r.item.Terms))
	for _, item := range r.item.Terms {
		item := item
		out = append(out, &termAssignmentResolver{item: item})
	}
	return out
}

func (r *contentResolver) Template() string {
	return r.item.Template
}

func (r *contentResolver) Metadata(ctx context.Context) (JSONValue, error) {
	projection, err := r.projection(ctx)
	if err != nil {
		return JSONValue{}, err
	}
	return JSONValue{Value: projection.Metadata}, nil
}

func (r *contentResolver) Links() JSONValue {
	return JSONValue{Value: map[string]any{"rest": resourcePath(r.item)}}
}

func (r *contentResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.item.CreatedAt}
}

func (r *contentResolver) UpdatedAt() graphql.Time {
	return graphql.Time{Time: r.item.UpdatedAt}
}

func (r *contentResolver) PublishedAt() *graphql.Time {
	return timePointer(r.item.PublishedAt)
}

func (r *contentResolver) DeletedAt() *graphql.Time {
	return timePointer(r.item.DeletedAt)
}

func (r *contentResolver) projection(ctx context.Context) (plugins.ContentProjection, error) {
	projection := plugins.ContentProjection{
		Slug:    localizedJSON(r.item.Slug),
		Title:   localizedJSON(r.item.Title),
		Content: localizedJSON(r.item.Body),
		Excerpt: localizedJSON(r.item.Excerpt),
	}
	allowedMetadata := r.item.Metadata.Public()
	if r.registry != nil || r.metaRegistry != nil {
		allowedMetadata = metadataProjection(r.metaRegistry, r.item, r.includePrivate)
	} else if r.includePrivate {
		allowedMetadata = r.item.Metadata
	}
	projection.Metadata = metadataJSON(allowedMetadata)
	filteredMetadata, err := plugins.FilterValue(ctx, r.registry, "content.metadata.public.filter", plugins.HookContext{
		Surface:       plugins.SurfaceREST,
		Path:          "/go-graphql",
		Locale:        locale.From(ctx),
		Principal:     stateFromContext(ctx).principal,
		Authenticated: stateFromContext(ctx).authenticated,
		Metadata: map[string]any{
			"resource":        "content-metadata",
			"include_private": r.includePrivate,
			"content_id":      string(r.item.ID),
		},
	}, projection.Metadata)
	if err != nil {
		return plugins.ContentProjection{}, err
	}
	projection.Metadata = sanitizeFilteredMetadata(filteredMetadata, allowedMetadata)
	filtered, err := plugins.FilterValue(ctx, r.registry, "graphql.content.filter", plugins.HookContext{
		Surface:       plugins.SurfacePublic,
		Path:          "/go-graphql",
		Locale:        locale.From(ctx),
		Principal:     stateFromContext(ctx).principal,
		Authenticated: stateFromContext(ctx).authenticated,
		Metadata: map[string]any{
			"resource":        "content",
			"include_private": r.includePrivate,
			"content_id":      string(r.item.ID),
		},
	}, projection)
	if err != nil {
		return plugins.ContentProjection{}, err
	}
	filtered.Metadata = sanitizeFilteredMetadata(filtered.Metadata, allowedMetadata)
	return filtered, nil
}

type termAssignmentResolver struct {
	item domaincontent.TermRef
}

func (r *termAssignmentResolver) Taxonomy() string {
	return r.item.Taxonomy
}

func (r *termAssignmentResolver) TermID() string {
	return r.item.TermID
}

type contentTypeResolver struct {
	item domaincontenttype.Type
}

func (r *contentTypeResolver) ID() string {
	return string(r.item.ID)
}

func (r *contentTypeResolver) Label() string {
	return r.item.Label
}

func (r *contentTypeResolver) Public() bool {
	return r.item.Public
}

func (r *contentTypeResolver) RestVisible() bool {
	return r.item.RESTVisible
}

func (r *contentTypeResolver) GraphqlVisible() bool {
	return r.item.GraphQLVisible
}

func (r *contentTypeResolver) Archive() bool {
	return r.item.Archive
}

func (r *contentTypeResolver) Permalink() string {
	return r.item.Permalink
}

func (r *contentTypeResolver) Supports() *contentTypeSupportsResolver {
	return &contentTypeSupportsResolver{item: r.item.Supports}
}

type contentTypeSupportsResolver struct {
	item domaincontenttype.Supports
}

func (r *contentTypeSupportsResolver) Title() bool         { return r.item.Title }
func (r *contentTypeSupportsResolver) Editor() bool        { return r.item.Editor }
func (r *contentTypeSupportsResolver) Excerpt() bool       { return r.item.Excerpt }
func (r *contentTypeSupportsResolver) FeaturedMedia() bool { return r.item.FeaturedMedia }
func (r *contentTypeSupportsResolver) Revisions() bool     { return r.item.Revisions }
func (r *contentTypeSupportsResolver) Taxonomies() bool    { return r.item.Taxonomies }
func (r *contentTypeSupportsResolver) CustomFields() bool  { return r.item.CustomFields }
func (r *contentTypeSupportsResolver) Comments() bool      { return r.item.Comments }

type taxonomyResolver struct {
	item domaintaxonomy.Definition
}

func (r *taxonomyResolver) Type() string  { return string(r.item.Type) }
func (r *taxonomyResolver) Label() string { return r.item.Label }
func (r *taxonomyResolver) Mode() string  { return string(r.item.Mode) }
func (r *taxonomyResolver) AssignedToKinds() []string {
	out := make([]string, 0, len(r.item.AssignedToKinds))
	for _, kind := range r.item.AssignedToKinds {
		out = append(out, string(kind))
	}
	return out
}
func (r *taxonomyResolver) Public() bool         { return r.item.Public }
func (r *taxonomyResolver) RestVisible() bool    { return r.item.RESTVisible }
func (r *taxonomyResolver) GraphqlVisible() bool { return r.item.GraphQLVisible }

type termResolver struct {
	item domaintaxonomy.Term
}

func (r *termResolver) ID() graphql.ID { return graphql.ID(r.item.ID) }
func (r *termResolver) Type() string   { return string(r.item.Type) }
func (r *termResolver) Name() JSONValue {
	return JSONValue{Value: map[string]string(r.item.Name)}
}
func (r *termResolver) Slug() JSONValue {
	return JSONValue{Value: map[string]string(r.item.Slug)}
}
func (r *termResolver) Description() JSONValue {
	return JSONValue{Value: map[string]string(r.item.Description)}
}
func (r *termResolver) ParentID() *string {
	if r.item.ParentID == "" {
		return nil
	}
	value := string(r.item.ParentID)
	return &value
}

type mediaResolver struct {
	item domainmedia.Asset
}

func (r *mediaResolver) ID() graphql.ID      { return graphql.ID(r.item.ID) }
func (r *mediaResolver) Filename() string    { return r.item.Filename }
func (r *mediaResolver) MimeType() string    { return r.item.MimeType }
func (r *mediaResolver) SizeBytes() string   { return strconv.FormatInt(r.item.SizeBytes, 10) }
func (r *mediaResolver) Width() int32        { return int32(r.item.Width) }
func (r *mediaResolver) Height() int32       { return int32(r.item.Height) }
func (r *mediaResolver) AltText() string     { return r.item.AltText }
func (r *mediaResolver) Caption() string     { return r.item.Caption }
func (r *mediaResolver) PublicURL() string   { return r.item.PublicURL }
func (r *mediaResolver) Metadata() JSONValue { return JSONValue{Value: r.item.PublicMeta} }
func (r *mediaResolver) Variants() []*mediaVariantResolver {
	out := make([]*mediaVariantResolver, 0, len(r.item.Variants))
	for _, item := range r.item.Variants {
		item := item
		out = append(out, &mediaVariantResolver{item: item})
	}
	return out
}
func (r *mediaResolver) CreatedAt() graphql.Time { return graphql.Time{Time: r.item.CreatedAt} }
func (r *mediaResolver) UpdatedAt() graphql.Time { return graphql.Time{Time: r.item.UpdatedAt} }

type mediaVariantResolver struct {
	item domainmedia.Variant
}

func (r *mediaVariantResolver) Name() string  { return r.item.Name }
func (r *mediaVariantResolver) URL() string   { return r.item.URL }
func (r *mediaVariantResolver) Width() int32  { return int32(r.item.Width) }
func (r *mediaVariantResolver) Height() int32 { return int32(r.item.Height) }

type authorResolver struct {
	profile domainusers.AuthorProfile
	media   appmedia.Service
}

func (r *authorResolver) ID() graphql.ID      { return graphql.ID(r.profile.ID) }
func (r *authorResolver) Slug() string        { return r.profile.Slug }
func (r *authorResolver) DisplayName() string { return r.profile.DisplayName }
func (r *authorResolver) Bio() string         { return r.profile.Bio }
func (r *authorResolver) AvatarMediaId() string {
	return r.profile.AvatarMediaID
}
func (r *authorResolver) AvatarURL(ctx context.Context) string {
	if mid := strings.TrimSpace(r.profile.AvatarMediaID); mid != "" {
		asset, ok, err := r.media.Get(ctx, domainmedia.ID(mid))
		if err == nil && ok {
			return asset.PublicURL
		}
	}
	return r.profile.AvatarURL
}
func (r *authorResolver) WebsiteURL() string { return r.profile.WebsiteURL }

type menuResolver struct {
	item domainmenus.Menu
}

func (r *menuResolver) ID() graphql.ID   { return graphql.ID(r.item.ID) }
func (r *menuResolver) Name() string     { return r.item.Name }
func (r *menuResolver) Location() string { return string(r.item.Location) }
func (r *menuResolver) Items() []*menuItemResolver {
	out := make([]*menuItemResolver, 0, len(r.item.Items))
	for _, item := range r.item.Items {
		item := item
		out = append(out, &menuItemResolver{item: item})
	}
	return out
}

type menuItemResolver struct {
	item domainmenus.Item
}

func (r *menuItemResolver) ID() graphql.ID   { return graphql.ID(r.item.ID) }
func (r *menuItemResolver) Label() string    { return r.item.Label }
func (r *menuItemResolver) URL() string      { return r.item.URL }
func (r *menuItemResolver) Kind() string     { return r.item.Kind }
func (r *menuItemResolver) TargetID() string { return r.item.TargetID }
func (r *menuItemResolver) Children() []*menuItemResolver {
	out := make([]*menuItemResolver, 0, len(r.item.Children))
	for _, item := range r.item.Children {
		item := item
		out = append(out, &menuItemResolver{item: item})
	}
	return out
}

type settingResolver struct {
	item domainsettings.Value
}

func (r *settingResolver) Key() string      { return string(r.item.Key) }
func (r *settingResolver) Value() JSONValue { return JSONValue{Value: r.item.Value} }
func (r *settingResolver) Public() bool     { return r.item.Public }

func buildContentQuery(args contentListArgs, kind domaincontent.Kind, publicOnly bool) (domaincontent.Query, error) {
	page := int(args.Page)
	if page <= 0 {
		page = 1
	}
	perPage := int(args.PerPage)
	if perPage <= 0 {
		perPage = 20
	}
	query := domaincontent.Query{
		PublicOnly: publicOnly,
		Page:       page,
		PerPage:    perPage,
		Locale:     stringValue(args.Locale),
		AuthorID:   stringValue(args.Author),
		Search:     stringValue(args.Search),
	}
	if kind != "" {
		query.Kinds = []domaincontent.Kind{kind}
	}
	if args.Status != nil {
		for _, status := range *args.Status {
			normalized := domaincontent.Status(strings.TrimSpace(status))
			if err := domaincontent.ValidateStatus(normalized); err != nil {
				return domaincontent.Query{}, err
			}
			query.Statuses = append(query.Statuses, normalized)
		}
	}
	if args.Taxonomy != nil {
		parts := strings.SplitN(strings.TrimSpace(*args.Taxonomy), ":", 2)
		query.Taxonomy = strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			query.TermID = strings.TrimSpace(parts[1])
		}
	}
	if args.Term != nil && strings.TrimSpace(*args.Term) != "" {
		query.TermID = strings.TrimSpace(*args.Term)
	}
	if args.After != nil {
		after := args.After.Time
		query.After = &after
	}
	if args.Before != nil {
		before := args.Before.Time
		query.Before = &before
	}
	if args.Sort != nil && strings.TrimSpace(*args.Sort) != "" {
		query.SortBy = domaincontent.SortField(strings.TrimSpace(*args.Sort))
	}
	query.SortDesc = strings.EqualFold(stringValue(args.Order), "desc")
	return query, nil
}

func localizedFromJSON(value JSONValue, allowEmpty bool) domaincontent.LocalizedText {
	result := domaincontent.LocalizedText{}
	switch raw := value.Value.(type) {
	case map[string]any:
		for key, entry := range raw {
			result[key] = fmt.Sprint(entry)
		}
	case map[string]string:
		for key, entry := range raw {
			result[key] = entry
		}
	case string:
		result["en"] = raw
	}
	if len(result) == 0 && !allowEmpty {
		result["en"] = ""
	}
	return result
}

func optionalLocalized(value *JSONValue) domaincontent.LocalizedText {
	if value == nil {
		return nil
	}
	return localizedFromJSON(*value, true)
}

func metadataFromJSON(value *JSONValue) domaincontent.Metadata {
	if value == nil {
		return nil
	}
	raw, ok := value.Value.(map[string]any)
	if !ok {
		return nil
	}
	metadata := make(domaincontent.Metadata, len(raw))
	for key, entry := range raw {
		metadata[key] = domaincontent.MetaValue{Value: entry, Public: true}
	}
	return metadata
}

func metadataJSON(metadata domaincontent.Metadata) map[string]any {
	if len(metadata) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(metadata))
	for key, value := range metadata {
		out[key] = value.Value
	}
	return out
}

func metadataProjection(metaRegistry *appmeta.Registry, entry domaincontent.Entry, includePrivate bool) domaincontent.Metadata {
	if metaRegistry == nil {
		if includePrivate {
			return entry.Metadata
		}
		return entry.Metadata.Public()
	}
	return metaRegistry.PublicMetadata(entry.Kind, entry.Metadata, includePrivate)
}

func localizedJSON(value domaincontent.LocalizedText) map[string]string {
	out := make(map[string]string, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}

func sanitizeFilteredMetadata(filtered map[string]any, allowed domaincontent.Metadata) map[string]any {
	if len(filtered) == 0 || len(allowed) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(filtered))
	for key := range allowed {
		value, ok := filtered[key]
		if !ok {
			continue
		}
		out[key] = value
	}
	return out
}

func termRefsFromInput(input *[]termRefInput) []domaincontent.TermRef {
	if input == nil {
		return nil
	}
	return termRefsFromSlice(*input)
}

func termRefsFromSlice(input []termRefInput) []domaincontent.TermRef {
	if len(input) == 0 {
		return nil
	}
	out := make([]domaincontent.TermRef, 0, len(input))
	for _, item := range input {
		out = append(out, domaincontent.TermRef{Taxonomy: item.Taxonomy, TermID: item.TermID})
	}
	return out
}

func menuItemsFromInput(input []menuItemInput) []domainmenus.Item {
	if len(input) == 0 {
		return nil
	}
	out := make([]domainmenus.Item, 0, len(input))
	for _, item := range input {
		out = append(out, domainmenus.Item{
			ID:       domainmenus.ItemID(item.ID),
			Label:    item.Label,
			URL:      item.URL,
			Kind:     stringValue(item.Kind),
			TargetID: stringValue(item.TargetID),
			Children: menuItemsFromPointer(item.Children),
		})
	}
	return out
}

func menuItemsFromPointer(input *[]menuItemInput) []domainmenus.Item {
	if input == nil {
		return nil
	}
	return menuItemsFromInput(*input)
}

func principalFromContext(ctx context.Context) domainauthz.Principal {
	return stateFromContext(ctx).principal
}

func publicOnly(ctx context.Context) bool {
	return !principalFromContext(ctx).Has(domainauthz.CapabilityContentReadPrivate)
}

func requirePrincipal(ctx context.Context) (domainauthz.Principal, error) {
	state := stateFromContext(ctx)
	if !state.authenticated {
		return domainauthz.Principal{}, fmt.Errorf("authentication is required")
	}
	return state.principal, nil
}

func hidesExistence(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not public") || strings.Contains(message, "not found")
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func timePointer(value *time.Time) *graphql.Time {
	if value == nil {
		return nil
	}
	out := graphql.Time{Time: *value}
	return &out
}

func resourcePath(entry domaincontent.Entry) string {
	switch entry.Kind {
	case domaincontent.KindPage:
		return "/go-json/go/v2/pages/" + string(entry.ID)
	default:
		return "/go-json/go/v2/posts/" + string(entry.ID)
	}
}
