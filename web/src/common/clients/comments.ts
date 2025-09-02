// Owner: feature/comments-system
// Dependencies: posts feature types

// Uncomment these imports when implementing the client
// import { createClient } from '@connectrpc/connect'
// import { createConnectTransport } from '@connectrpc/connect-web'

// === SERVICE CLIENT START ===
// Add service client initialization below this line
// Follow pattern from posts.ts

// const transport = createConnectTransport({
//   baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
// })

// export const commentsClient = createClient(
//   CommentService,
//   transport
// )

// === SERVICE CLIENT END ===

// === HOOKS START ===
// Add React Query hooks below this line

// export function useComments(postId: string) {
//   return useQuery({
//     queryKey: ['comments', postId],
//     queryFn: async () => {
//       // Implementation
//     },
//   })
// }

// export function useAddComment() {
//   return useMutation({
//     mutationFn: async (input: AddCommentInput) => {
//       // Implementation
//     },
//   })
// }

// === HOOKS END ===
