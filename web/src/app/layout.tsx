import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Alunalun',
  description: 'Location-based social platform',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        {/* === PROVIDER WRAPPERS START === */}
        {/* Each feature adds their providers in their section */}
        {/* Maintain nesting order when merging */}
        
        {/* [AUTH-PROVIDER-START] */}
        {/* Owner: feature/auth */}
        {/* <AuthProvider> */}
        {/* [AUTH-PROVIDER-END] */}
        
        {/* [QUERY-PROVIDER-START] */}
        {/* Owner: base/main */}
        {/* <QueryClientProvider client={queryClient}> */}
        
          {/* [LOCATION-PROVIDER-START] */}
          {/* Owner: feature/location-feed or feature/map-pins */}
          {/* <LocationProvider> */}
          {/* [LOCATION-PROVIDER-END] */}
          
          {/* [THEME-PROVIDER-START] */}
          {/* Owner: feature/theme */}
          {/* <ThemeProvider> */}
          {/* [THEME-PROVIDER-END] */}
          
          {children}
          
          {/* [THEME-PROVIDER-CLOSE] */}
          {/* </ThemeProvider> */}
          {/* [LOCATION-PROVIDER-CLOSE] */}
          {/* </LocationProvider> */}
          
        {/* [QUERY-PROVIDER-CLOSE] */}
        {/* </QueryClientProvider> */}
        {/* [AUTH-PROVIDER-CLOSE] */}
        {/* </AuthProvider> */}
        
        {/* === PROVIDER WRAPPERS END === */}
      </body>
    </html>
  )
}