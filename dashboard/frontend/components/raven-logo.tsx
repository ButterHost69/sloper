type RavenLogoProps = {
  size?: number;
  className?: string;
};

export function RavenLogo({ size = 32, className }: RavenLogoProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 48 48"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      <path
        d="M12.2 20.8C9.7 13.1 14.6 6.3 22.7 5.5c2.5-.2 4.2.7 5.6 1.7 1.5-1.3 3.2-2 5.6-1.7 7.6.8 11 7.8 8.7 15.4 1.4 4.6-.2 10.2-4.3 14.2-3.2 3.2-7 5.2-10 5.2s-6.8-2-10-5.2c-4.1-4-5.7-9.7-4.1-14.3Z"
        fill="#252529"
        stroke="#111114"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <path
        d="M14.1 19.3c-1.4-4.7 1.8-9.2 6.6-10.1-1.5 2.4-1.6 5.1-.5 7.5-2.1-.3-4.2.8-6.1 2.6Z"
        fill="#5E5CE6"
        opacity=".9"
      />
      <path
        d="M33.9 19.3c1.4-4.7-1.8-9.2-6.6-10.1 1.5 2.4 1.6 5.1.5 7.5 2.1-.3 4.2.8 6.1 2.6Z"
        fill="#2997FF"
        opacity=".82"
      />
      <path
        d="M10.6 27.4c-2.1 1.4-3.5 3.5-3.7 5.8 2.6-.2 5-1.3 6.7-3.2M37.4 27.4c2.1 1.4 3.5 3.5 3.7 5.8-2.6-.2-5-1.3-6.7-3.2"
        stroke="#111114"
        strokeWidth="2.4"
        strokeLinecap="round"
      />
      <ellipse cx="18.4" cy="23.2" rx="5.1" ry="5.7" fill="#F5F5F7" />
      <ellipse cx="29.6" cy="23.2" rx="5.1" ry="5.7" fill="#F5F5F7" />
      <ellipse cx="19.2" cy="24" rx="2.2" ry="2.8" fill="#111114" />
      <ellipse cx="28.8" cy="24" rx="2.2" ry="2.8" fill="#111114" />
      <circle cx="19.9" cy="22.9" r=".85" fill="#fff" />
      <circle cx="29.5" cy="22.9" r=".85" fill="#fff" />
      <path d="m24 27.1 6.1 4.2-6.1 4.3-6.1-4.3 6.1-4.2Z" fill="#F5A623" stroke="#111114" strokeWidth="1.2" strokeLinejoin="round" />
      <path d="M24 31.4v5.1M20.4 40.1 17.8 43M27.6 40.1l2.6 2.9" stroke="#F5A623" strokeWidth="2" strokeLinecap="round" />
      <path d="M15.3 31.2c1.8 4.2 4.8 6.5 8.7 6.5s6.9-2.3 8.7-6.5" stroke="#111114" strokeWidth="1.4" strokeLinecap="round" opacity=".8" />
    </svg>
  );
}
