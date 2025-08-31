// Owner: feature/map-pins
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

// export const mapClient = createClient(
//   MapPinService,
//   transport
// )

// === SERVICE CLIENT END ===

// === HOOKS START ===
// Add React Query hooks below this line

// export function useMapPins(bounds: MapBounds) {
//   return useQuery({
//     queryKey: ['map-pins', bounds],
//     queryFn: async () => {
//       // Implementation
//     },
//   })
// }

// export function usePinDetails(pinId: string) {
//   return useQuery({
//     queryKey: ['pin-details', pinId],
//     queryFn: async () => {
//       // Implementation
//     },
//   })
// }

// === HOOKS END ===