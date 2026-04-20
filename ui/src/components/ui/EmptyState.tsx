import React from "react";

export const EmptyState = ({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon: React.ElementType;
  title: string;
  description?: string;
  action?: React.ReactNode;
}) => (
  <div className="flex flex-col items-center justify-center py-20 gap-4">
    <div className="w-14 h-14 rounded-full bg-[#1f2128] border border-[#2c3235] flex items-center justify-center">
      <Icon size={24} className="text-[#34383e]" />
    </div>
    <div className="text-center space-y-1">
      <p className="text-[#d9d9d9] font-medium text-sm">{title}</p>
      {description && (
        <p className="text-[#8e9193] text-xs max-w-xs">{description}</p>
      )}
    </div>
    {action && <div className="mt-2">{action}</div>}
  </div>
);
