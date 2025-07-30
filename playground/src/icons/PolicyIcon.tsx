// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React from 'react';

interface IconProps {
  width?: number;
  height?: number;
  className?: string;
}

export const PolicyIcon: React.FC<IconProps> = ({ 
  width = 16, 
  height = 16, 
  className 
}) => (
  <svg 
    width={width} 
    height={height} 
    viewBox="0 0 16 16" 
    fill="currentColor"
    className={className}
    aria-hidden="true"
  >
    <path d="M8 1a2 2 0 0 1 2 2v4H6V3a2 2 0 0 1 2-2zM3 6a1 1 0 0 0-1 1v3a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7a1 1 0 0 0-1-1H3z"/>
  </svg>
);