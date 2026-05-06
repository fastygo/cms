package graphqlplugin

const schema = `
scalar JSON
scalar Time

schema {
  query: Query
  mutation: Mutation
}

type Query {
  posts(page: Int = 1, perPage: Int = 20, status: [String!], author: String, taxonomy: String, term: String, search: String, locale: String, after: Time, before: Time, sort: String, order: String): ContentList!
  post(id: ID, slug: String, locale: String): Content
  pages(page: Int = 1, perPage: Int = 20, status: [String!], author: String, taxonomy: String, term: String, search: String, locale: String, after: Time, before: Time, sort: String, order: String): ContentList!
  page(id: ID, slug: String, locale: String): Content
  contentTypes: [ContentType!]!
  taxonomies: [Taxonomy!]!
  terms(type: String!): [Term!]!
  media(id: ID): [Media!]!
  authors: [Author!]!
  menus(location: String): [Menu!]!
  settings: [Setting!]!
  search(query: String!, page: Int = 1, perPage: Int = 20, locale: String): ContentList!
}

type Mutation {
  createPost(input: ContentInput!): Content!
  createPage(input: ContentInput!): Content!
  updateContent(id: ID!, input: ContentInput!): Content!
  publishContent(id: ID!): Content!
  scheduleContent(id: ID!, publishedAt: Time!): Content!
  trashContent(id: ID!): Content!
  restoreContent(id: ID!): Content!
  assignTerms(contentID: ID!, terms: [TermRefInput!]!): Content!
  attachFeaturedMedia(contentID: ID!, assetID: ID!): Content!
  saveMenu(input: MenuInput!): Menu!
  saveSetting(input: SettingInput!): Setting!
}

type Pagination {
  page: Int!
  perPage: Int!
  total: Int!
  totalPages: Int!
}

type ContentList {
  items: [Content!]!
  pagination: Pagination!
}

type Content {
  id: ID!
  kind: String!
  status: String!
  visibility: String!
  slug: JSON!
  title: JSON!
  content: JSON!
  excerpt: JSON!
  authorID: String!
  featuredMediaID: String
  taxonomies: [TermAssignment!]!
  template: String!
  metadata: JSON!
  links: JSON!
  createdAt: Time!
  updatedAt: Time!
  publishedAt: Time
  deletedAt: Time
}

type TermAssignment {
  taxonomy: String!
  termID: String!
}

type ContentType {
  id: String!
  label: String!
  public: Boolean!
  restVisible: Boolean!
  graphqlVisible: Boolean!
  archive: Boolean!
  permalink: String!
  supports: ContentTypeSupports!
}

type ContentTypeSupports {
  title: Boolean!
  editor: Boolean!
  excerpt: Boolean!
  featuredMedia: Boolean!
  revisions: Boolean!
  taxonomies: Boolean!
  customFields: Boolean!
  comments: Boolean!
}

type Taxonomy {
  type: String!
  label: String!
  mode: String!
  assignedToKinds: [String!]!
  public: Boolean!
  restVisible: Boolean!
  graphqlVisible: Boolean!
}

type Term {
  id: ID!
  type: String!
  name: JSON!
  slug: JSON!
  description: JSON!
  parentID: String
}

type Media {
  id: ID!
  filename: String!
  mimeType: String!
  sizeBytes: String!
  width: Int!
  height: Int!
  altText: String!
  caption: String!
  publicURL: String!
  metadata: JSON!
  variants: [MediaVariant!]!
  createdAt: Time!
  updatedAt: Time!
}

type MediaVariant {
  name: String!
  url: String!
  width: Int!
  height: Int!
}

type Author {
  id: ID!
  slug: String!
  displayName: String!
  bio: String!
  avatarURL: String!
  websiteURL: String!
}

type Menu {
  id: ID!
  name: String!
  location: String!
  items: [MenuItem!]!
}

type MenuItem {
  id: ID!
  label: String!
  url: String!
  kind: String!
  targetID: String!
  children: [MenuItem!]!
}

type Setting {
  key: String!
  value: JSON!
  public: Boolean!
}

input ContentInput {
  status: String
  title: JSON!
  slug: JSON!
  content: JSON!
  excerpt: JSON
  authorID: String
  featuredMediaID: String
  template: String
  metadata: JSON
  terms: [TermRefInput!]
  publishedAt: Time
}

input TermRefInput {
  taxonomy: String!
  termID: String!
}

input MenuInput {
  id: String!
  name: String!
  location: String!
  items: [MenuItemInput!]!
}

input MenuItemInput {
  id: String!
  label: String!
  url: String!
  kind: String
  targetID: String
  children: [MenuItemInput!]
}

input SettingInput {
  key: String!
  value: JSON!
  public: Boolean = false
}
`
