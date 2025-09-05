import { Map } from "@/features/map";

export default function HomePage() {
  return (
    <div className="h-screen w-screen bg-gray-100">
      <Map className="h-full w-full" />
    </div>
  );
}
