// Main landing page
// Each feature can add their components here

export default function HomePage() {
  return (
    <main className="min-h-screen">
      {/* === PAGE SECTIONS START === */}
      {/* Each feature adds their main components in their section */}
      
      {/* [MAP-SECTION-START] */}
      {/* Owner: feature/map-pins */}
      {/* <MapContainer /> */}
      {/* [MAP-SECTION-END] */}
      
      {/* [FEED-SECTION-START] */}
      {/* Owner: feature/location-feed */}
      {/* <FeedContainer /> */}
      {/* [FEED-SECTION-END] */}
      
      {/* [POST-FORM-SECTION-START] */}
      {/* Owner: feature/posts-core */}
      {/* <PostFormFloating /> */}
      {/* [POST-FORM-SECTION-END] */}
      
      {/* === PAGE SECTIONS END === */}
      
      {/* Temporary placeholder content */}
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <h1 className="text-4xl font-bold mb-4">Alunalun</h1>
          <p className="text-gray-600">Location-based social platform</p>
          <p className="text-sm text-gray-400 mt-8">
            Features will be added by parallel development branches
          </p>
        </div>
      </div>
    </main>
  )
}