import Image from "next/image";

import { Button } from "@/shared/ui/button";

export function VkButton() {
  return (
    <Button type="button" variant="outline" className="w-full">
      <Image
        src="/assets/vk.png"
        alt=""
        width={17}
        height={16}
        className="h-4 w-[17px]"
      />
      ВКонтакте
    </Button>
  );
}
