// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

use std::collections::HashMap;
use std::sync::RwLock;
use std::error::Error;
use std::fmt;
use serde::Deserialize;
use lens_sdk::StreamOption;
use lens_sdk::error::LensError;

lens_sdk::define!(PARAMETERS: Parameters, try_transform);

#[derive(Clone, PartialEq, Eq, PartialOrd, Ord, Debug, Hash)]
enum ModuleError {
    PropertyNotFoundError{requested: String},
}

impl Error for ModuleError { }

impl fmt::Display for ModuleError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match &*self {
            ModuleError::PropertyNotFoundError { requested } =>
                write!(f, "The requested property was not found. Requested: {}", requested),
        }
    }
}

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub src: String,
    pub dst: String,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);

fn try_transform(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<HashMap<String, serde_json::Value>>>>,
) -> Result<StreamOption<HashMap<String, serde_json::Value>>, Box<dyn Error>> {
    let params = PARAMETERS.read()?
        .clone()
        .ok_or(LensError::ParametersNotSetError)?;

    for item in iter {
        let mut input = match item? {
            Some(v) => v,
            None => return Ok(StreamOption::None),
        };

        let value = input.get_mut(&params.src)
            .ok_or(ModuleError::PropertyNotFoundError{requested: params.src.clone()})?
            .clone();

        input.insert(params.dst, value);

        return Ok(StreamOption::Some(input))
    }

    Ok(StreamOption::EndOfStream)
}
