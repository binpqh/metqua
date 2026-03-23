import pandas as pd
import numpy as np
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder

dataframe = pd.read_csv('student_performance.csv')
X = dataframe.drop("Passed", axis=1)
y = dataframe["Passed"]

# Handle dummy variables trap by dropping the first category in each feature
encoder = OneHotEncoder(drop="first", handle_unknown="ignore")

column_transformer = ColumnTransformer(
    transformers=[
        ('encoder', encoder, ['Gender'])
    ])
X1 = np.array(column_transformer.fit_transform(X))
print(X1)