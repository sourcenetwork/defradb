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

lens_sdk::define!(PARAMETERS: Parameters, try_transform, try_inverse);

#[derive(Clone, PartialEq, Eq, PartialOrd, Ord, Debug, Hash)]
enum ModuleError {
    PropertyNotFoundError{requested: String},
    PropertyNotNumberError{requested: String},
    InvalidIncrementValueError,
}

impl Error for ModuleError { }

impl fmt::Display for ModuleError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match &*self {
            ModuleError::PropertyNotFoundError { requested } =>
                write!(f, "The requested property was not found. Requested: {}", requested),
            ModuleError::PropertyNotNumberError { requested } =>
                write!(f, "The requested property is not a number. Requested: {}", requested),
            ModuleError::InvalidIncrementValueError =>
                write!(f, "The increment value must be a number"),
        }
    }
}

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub field: String,
    pub value: serde_json::Value,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);

fn try_transform(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<HashMap<String, serde_json::Value>>>>,
) -> Result<StreamOption<HashMap<String, serde_json::Value>>, Box<dyn Error>> {
    apply_increment(iter, |current, increment| current + increment)
}

fn try_inverse(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<HashMap<String, serde_json::Value>>>>,
) -> Result<StreamOption<HashMap<String, serde_json::Value>>, Box<dyn Error>> {
    apply_increment(iter, |current, increment| current - increment)
}

fn apply_increment(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<HashMap<String, serde_json::Value>>>>,
    operation: fn(i64, i64) -> i64,
) -> Result<StreamOption<HashMap<String, serde_json::Value>>, Box<dyn Error>> {
    let params = PARAMETERS.read()?
        .clone()
        .ok_or(LensError::ParametersNotSetError)?;

    let increment = params.value.as_i64()
        .ok_or(ModuleError::InvalidIncrementValueError)?;

    for item in iter {
        let mut input = match item? {
            Some(v) => v,
            None => return Ok(StreamOption::None),
        };

        let field_value = input.get_mut(&params.field)
            .ok_or(ModuleError::PropertyNotFoundError{requested: params.field.clone()})?;

        let current_value = field_value.as_i64()
            .ok_or(ModuleError::PropertyNotNumberError{requested: params.field.clone()})?;

        let new_value = operation(current_value, increment);

        input.insert(params.field.clone(), serde_json::Value::Number(new_value.into()));

        return Ok(StreamOption::Some(input))
    }

    Ok(StreamOption::EndOfStream)
}
