// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

use std::collections::HashMap;
use std::sync::RwLock;
use std::error::Error;
use serde::Deserialize;
use lens_sdk::StreamOption;
use lens_sdk::error::LensError;

lens_sdk::define!(PARAMETERS: Parameters, try_transform);

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub values: Vec<HashMap<String, serde_json::Value>>,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);
static PARAM_INDEX: RwLock<usize> = RwLock::new(0);

fn try_transform(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<HashMap<String, serde_json::Value>>>>,
) -> Result<StreamOption<HashMap<String, serde_json::Value>>, Box<dyn Error>> {
    let params = PARAMETERS.read()?
        .clone()
        .ok_or(LensError::ParametersNotSetError)?;

    let param_index = PARAM_INDEX.read()?
        .clone();

    if param_index < params.values.len() {
        let result = &params.values[param_index];

        let mut dst = PARAM_INDEX.write()?;
        *dst = param_index+1;

        return Ok(StreamOption::Some(result.clone()))
    }

    for item in iter {
        let input = match item? {
            Some(v) => v,
            None => return Ok(StreamOption::None),
        };

        return Ok(StreamOption::Some(input))
    }

    Ok(StreamOption::EndOfStream)
}
