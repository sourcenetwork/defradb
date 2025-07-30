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

export const KeypairIcon: React.FC<IconProps> = ({ 
  width = 16, 
  height = 16, 
  className, 
}) => (
  <svg 
    width={width} 
    height={height} 
    viewBox="0 0 16 16" 
    fill="currentColor"
    className={className}
    aria-hidden="true"
  >
    <path d="M3.5 11.5a3.5 3.5 0 1 1 3.163-5H14L15.5 8 14 9.5l-1-1-1 1-1-1-1 1-1-1-1 1H6.663a3.5 3.5 0 0 1-3.163 2zM2.5 9a1 1 0 1 0 0-2 1 1 0 0 0 0 2z"/>
  </svg>
);