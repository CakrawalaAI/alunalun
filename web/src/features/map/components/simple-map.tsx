import { useRef, useEffect, useState } from 'react';
import maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';

interface SimpleMapProps {
  className?: string;
  center?: [number, number];
  zoom?: number;
}

export function SimpleMap({ 
  className = "", 
  center = [106.827, -6.175], 
  zoom = 12 
}: SimpleMapProps) {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<maplibregl.Map | null>(null);
  const initialized = useRef(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!mapContainer.current || initialized.current) return;

    try {
      console.log('Initializing MapLibre once...');
      initialized.current = true;
      
      map.current = new maplibregl.Map({
        container: mapContainer.current,
        style: 'https://tiles.openfreemap.org/styles/liberty',
        center: center,
        zoom: zoom,
        attributionControl: true,
      });

      map.current.on('load', () => {
        console.log('Map loaded successfully');
        setIsLoaded(true);
        setError(null);
        
        if (map.current) {
          map.current.addControl(new maplibregl.NavigationControl(), 'top-right');
        }
      });

      map.current.on('error', (e) => {
        console.error('MapLibre error:', e);
        setError(e.error?.message || 'Map failed to load');
      });

    } catch (err) {
      console.error('Failed to initialize map:', err);
      setError(err instanceof Error ? err.message : 'Failed to initialize map');
    }

    return () => {
      if (map.current) {
        console.log('Cleaning up map...');
        map.current.remove();
        map.current = null;
      }
      initialized.current = false;
      setIsLoaded(false);
      setError(null);
    };
  }, []); // Empty dependency array - initialize only once

  if (error) {
    return (
      <div className="flex h-full w-full items-center justify-center bg-red-50">
        <div className="text-center p-4">
          <h2 className="text-lg font-semibold text-red-600 mb-2">
            Map Error
          </h2>
          <p className="text-sm text-red-500">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Reload Page
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative w-full h-full">
      {!isLoaded && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-50 z-10">
          <div className="text-center">
            <div className="mb-4 inline-block h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600" />
            <p className="text-sm text-gray-600">Loading map...</p>
          </div>
        </div>
      )}
      <div 
        ref={mapContainer}
        className={`w-full h-full ${className}`}
        style={{
          minHeight: '400px',
          position: 'relative'
        }}
      />
    </div>
  );
}