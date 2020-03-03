<p>Packages:</p>
<ul>
<li>
<a href="#example.examples.gardener.cloud%2fv1beta1">example.examples.gardener.cloud/v1beta1</a>
</li>
</ul>
<h2 id="example.examples.gardener.cloud/v1beta1">example.examples.gardener.cloud/v1beta1</h2>
<p>
<p>Package v1beta1 contains example API resources.</p>
</p>
Resource Types:
<ul><li>
<a href="#example.examples.gardener.cloud/v1beta1.Example">Example</a>
</li></ul>
<h3 id="example.examples.gardener.cloud/v1beta1.Example">Example
</h3>
<p>
<p>Example is an example for a custom resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
example.examples.gardener.cloud/v1beta1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Example</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.16/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#example.examples.gardener.cloud/v1beta1.ExampleSpec">
ExampleSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>URL</code></br>
<em>
string
</em>
</td>
<td>
<p>URL is the address of the example</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="example.examples.gardener.cloud/v1beta1.ExampleSpec">ExampleSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#example.examples.gardener.cloud/v1beta1.Example">Example</a>)
</p>
<p>
<p>ExampleSpec is  the specification for an example object.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>URL</code></br>
<em>
string
</em>
</td>
<td>
<p>URL is the address of the example</p>
</td>
</tr>
</tbody>
</table>
<hr/>
