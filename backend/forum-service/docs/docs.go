// Package docs provides documentation for the forum service API
package docs

// ForumServiceAPI provides documentation for the forum service endpoints
const ForumServiceAPI = `
Forum Service API Documentation

The forum service provides endpoints for managing posts and comments in a forum system.

Authentication:
All endpoints except GetPosts and GetComments require a valid JWT token in the Authorization header.
Format: Authorization: Bearer <token>

Posts Endpoints:

1. Create Post
   POST /api/posts
   Request Body:
   {
     "title": string,
     "content": string
   }
   Response: 201 Created
   {
     "id": number,
     "title": string,
     "content": string,
     "author_id": number,
     "created_at": string (ISO 8601)
   }

2. Get Posts
   GET /api/posts
   Response: 200 OK
   {
     "posts": [
       {
         "id": number,
         "title": string,
         "content": string,
         "author_id": number,
         "author_name": string,
         "created_at": string (ISO 8601)
       }
     ]
   }

3. Update Post
   PUT /api/posts/{post_id}
   Request Body:
   {
     "title": string,
     "content": string
   }
   Response: 200 OK
   {
     "id": number,
     "title": string,
     "content": string,
     "author_id": number,
     "created_at": string (ISO 8601)
   }

4. Delete Post
   DELETE /api/posts/{post_id}
   Response: 204 No Content

Comments Endpoints:

1. Create Comment
   POST /api/posts/{post_id}/comments
   Request Body:
   {
     "content": string
   }
   Response: 201 Created
   {
     "id": number,
     "post_id": number,
     "content": string,
     "author_id": number,
     "created_at": string (ISO 8601)
   }

2. Get Comments
   GET /api/posts/{post_id}/comments
   Response: 200 OK
   {
     "comments": [
       {
         "id": number,
         "post_id": number,
         "content": string,
         "author_id": number,
         "author_name": string,
         "created_at": string (ISO 8601)
       }
     ]
   }

3. Update Comment
   PUT /api/posts/{post_id}/comments/{comment_id}
   Request Body:
   {
     "content": string
   }
   Response: 200 OK
   {
     "id": number,
     "post_id": number,
     "content": string,
     "author_id": number,
     "created_at": string (ISO 8601)
   }

4. Delete Comment
   DELETE /api/posts/{post_id}/comments/{comment_id}
   Response: 204 No Content

Error Responses:
All endpoints may return the following errors:

400 Bad Request
{
  "error": "Invalid request parameters"
}

401 Unauthorized
{
  "error": "Invalid or missing authentication token"
}

403 Forbidden
{
  "error": "Permission denied"
}

404 Not Found
{
  "error": "Resource not found"
}

500 Internal Server Error
{
  "error": "Internal server error"
}

Permissions:
- Regular users can create, update, and delete their own posts and comments
- Admin users can update and delete any posts and comments
- All users can view posts and comments
`

// ForumServiceGRPC provides documentation for the forum service gRPC endpoints
const ForumServiceGRPC = `
Forum Service gRPC API Documentation

The forum service provides gRPC endpoints for managing posts and comments.

Service: forum.ForumService

Posts Methods:

1. CreatePost
   Request: CreatePostRequest
   {
     token: string
     title: string
     content: string
   }
   Response: PostResponse
   {
     post: {
       id: int64
       title: string
       content: string
       author_id: int64
       created_at: string
     }
   }

2. GetPosts
   Request: GetPostsRequest
   {}
   Response: GetPostsResponse
   {
     posts: [
       {
         id: int64
         title: string
         content: string
         author_id: int64
         created_at: string
       }
     ]
     author_names: map<int64, string>
   }

3. UpdatePost
   Request: UpdatePostRequest
   {
     token: string
     post_id: int64
     title: string
     content: string
   }
   Response: PostResponse
   {
     post: {
       id: int64
       title: string
       content: string
       author_id: int64
       created_at: string
     }
   }

4. DeletePost
   Request: DeletePostRequest
   {
     token: string
     post_id: int64
   }
   Response: DeletePostResponse
   {}

Comments Methods:

1. CreateComment
   Request: CreateCommentRequest
   {
     token: string
     post_id: int64
     content: string
   }
   Response: CommentResponse
   {
     comment: {
       id: int64
       post_id: int64
       content: string
       author_id: int64
       created_at: string
     }
   }

2. GetComments
   Request: GetCommentsRequest
   {
     post_id: int64
   }
   Response: GetCommentsResponse
   {
     comments: [
       {
         id: int64
         post_id: int64
         content: string
         author_id: int64
         created_at: string
       }
     ]
     author_names: map<int64, string>
   }

3. UpdateComment
   Request: UpdateCommentRequest
   {
     token: string
     post_id: int64
     comment_id: int64
     content: string
   }
   Response: CommentResponse
   {
     comment: {
       id: int64
       post_id: int64
       content: string
       author_id: int64
       created_at: string
     }
   }

4. DeleteComment
   Request: DeleteCommentRequest
   {
     token: string
     post_id: int64
     comment_id: int64
   }
   Response: DeleteCommentResponse
   {}

Error Codes:
- INVALID_ARGUMENT (3): Invalid request parameters
- UNAUTHENTICATED (16): Invalid or missing authentication token
- PERMISSION_DENIED (7): User doesn't have permission to perform the action
- NOT_FOUND (5): Resource not found
- INTERNAL (13): Internal server error
`













